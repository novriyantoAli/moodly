package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/common/contract"
	"github.com/novriyantoAli/moodly/internal/application/payment/domain"
	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	"github.com/novriyantoAli/moodly/internal/application/payment/repository"
	"github.com/novriyantoAli/moodly/internal/pkg/database"
	"github.com/novriyantoAli/moodly/internal/pkg/jwt"
	"go.uber.org/zap"
	"gorm.io/gorm"

	billEntity "github.com/novriyantoAli/moodly/internal/application/bill/entity"
	billService "github.com/novriyantoAli/moodly/internal/application/bill/service"
	userService "github.com/novriyantoAli/moodly/internal/application/user/service"
)

type PaymentService interface {
	CreatePayment(
		ctx context.Context,
		req *dto.CreatePaymentRequest,
	) (*dto.PaymentResponse, error)

	GetPaymentByID(
		ctx context.Context,
		id uint,
	) (*dto.PaymentResponse, error)

	GetPaymentByNumber(
		ctx context.Context,
		paymentNumber string,
	) (*dto.PaymentResponse, error)

	GetPayments(
		ctx context.Context,
		filter *dto.PaymentFilter,
	) (*dto.PaymentListResponse, error)

	GetPaymentsByBillID(
		ctx context.Context,
		billID string,
	) ([]dto.PaymentResponse, error)

	UpdatePayment(
		ctx context.Context,
		id uint,
		req *dto.UpdatePaymentRequest,
	) (*dto.PaymentResponse, error)
}

type paymentService struct {
	repo      repository.PaymentRepository
	billSvc   billService.BillService
	userSvc   userService.UserService
	generator contract.PaymentNumberGenerator
	txManager database.TransactionManagerI
	logger    *zap.Logger
}

func NewPaymentService(
	repo repository.PaymentRepository,
	billSvc billService.BillService,
	userSvc userService.UserService,
	generator contract.PaymentNumberGenerator,
	txManager database.TransactionManagerI,
	logger *zap.Logger,
) PaymentService {
	return &paymentService{
		repo:      repo,
		billSvc:   billSvc,
		userSvc:   userSvc,
		generator: generator,
		txManager: txManager,
		logger:    logger,
	}
}

func GetClaims(ctx context.Context) (*jwt.Claims, error) {
	claims, ok := ctx.Value(jwt.ClaimsKey).(*jwt.Claims)
	if !ok || claims == nil {
		return nil, errors.New("claims not found in context")
	}
	return claims, nil
}

func (s *paymentService) CreatePayment(
	ctx context.Context,
	req *dto.CreatePaymentRequest,
) (*dto.PaymentResponse, error) {

	method := entity.PaymentMethod(req.Method)
	if !method.IsValid() {
		return nil, errors.New("invalid payment method")
	}

	var payment *entity.Payment

	err := s.txManager.WithinTransaction(
		ctx,
		func(ctx context.Context) error {

			// Validate bill
			bill, err := s.billSvc.LockForPayment(
				ctx,
				req.BillNumber,
			)
			if err != nil {
				return err
			}

			exists, err := s.repo.ExistsActivePaymentByBillID(
				ctx,
				bill.ID.String(),
			)
			if err != nil {
				return err
			}

			if exists {
				return errors.New(
					"active payment already exists for this bill",
				)
			}

			now := time.Now()

			payment = &entity.Payment{
				BillID:   bill.ID,
				Amount:   bill.Amount,
				Currency: "IDR",
				Method:   method,
				Status:   entity.PaymentStatusPending,
				Description: fmt.Sprintf(
					"Pembayaran %s",
					bill.BillNumber,
				),
				CreatedAt: now,
				UpdatedAt: now,
			}

			if method == entity.PaymentMethodCash {

				claim, err := GetClaims(ctx)
				if err != nil {
					return err
				}

				user, err := s.userSvc.GetUserByID(
					ctx,
					claim.UserID,
				)
				if err != nil {
					return err
				}

				payment.CreatedBy = &user.ID

				s.completePayment(payment)

				s.logger.Info(
					"Cash payment completed",
					zap.String(
						"payment_number",
						payment.PaymentNumber,
					),
				)
			}

			payment.PaymentNumber = s.generator.Generate()

			s.logger.Info(
				"Creating payment",
				zap.String(
					"bill_number",
					bill.BillNumber,
				),
				zap.String(
					"payment_number",
					payment.PaymentNumber,
				),
				zap.String(
					"method",
					string(payment.Method),
				),
			)

			if err := s.repo.Create(
				ctx,
				payment,
			); err != nil {
				return err
			}

			// Update bill bila payment langsung lunas
			if payment.Status == entity.PaymentStatusCompleted {
				if err := s.billSvc.MarkAsPaid(ctx, bill.ID.String()); err != nil {
					return err
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) GetPaymentByID(
	ctx context.Context,
	id uint,
) (*dto.PaymentResponse, error) {

	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}

		return nil, err
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) GetPaymentByNumber(
	ctx context.Context,
	paymentNumber string,
) (*dto.PaymentResponse, error) {

	payment, err := s.repo.GetByPaymentNumber(
		ctx,
		paymentNumber,
	)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}

		return nil, err
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) GetPaymentsByBillID(
	ctx context.Context,
	billID string,
) ([]dto.PaymentResponse, error) {

	payments, err := s.repo.GetByBillID(
		ctx,
		billID,
	)

	if err != nil {
		return nil, err
	}

	responses := make(
		[]dto.PaymentResponse,
		0,
		len(payments),
	)

	for _, payment := range payments {
		responses = append(
			responses,
			*s.entityToResponse(&payment),
		)
	}

	return responses, nil
}

func (s *paymentService) GetPayments(ctx context.Context, filter *dto.PaymentFilter) (*dto.PaymentListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	payments, totalCount, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, *s.entityToResponse(&payment))
	}

	return &dto.PaymentListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}
func (s *paymentService) UpdatePayment(
	ctx context.Context,
	id uint,
	req *dto.UpdatePaymentRequest,
) (*dto.PaymentResponse, error) {

	status := entity.PaymentStatus(req.Status)

	if !status.IsValid() {
		return nil, errors.New(
			"invalid payment status",
		)
	}

	var response *dto.PaymentResponse

	err := s.txManager.WithinTransaction(
		ctx,
		func(txCtx context.Context) error {

			payment, err := s.repo.GetByIDForUpdate(
				txCtx,
				id,
			)

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New(
						"payment not found",
					)
				}

				return err
			}

			// idempotent
			if payment.Status == status {
				response = s.entityToResponse(
					payment,
				)
				return nil
			}

			if domain.IsFinal(payment.Status) {
				return errors.New(
					"payment is in final state and cannot be modified",
				)
			}

			if !domain.IsValidTransition(payment.Status, status) {
				return errors.New(
					"invalid payment status transition",
				)
			}

			s.logger.Info(
				"Updating payment status",
				zap.Uint(
					"payment_id",
					payment.ID,
				),
				zap.String(
					"from",
					payment.Status.String(),
				),
				zap.String(
					"to",
					status.String(),
				),
			)

			payment.Status = status
			payment.Description = req.Description
			payment.UpdatedAt = time.Now()

			// otomatis isi PaidAt
			if status == entity.PaymentStatusCompleted && payment.PaidAt == nil {
				now := time.Now()
				payment.PaidAt = &now
			}

			if err := s.repo.Update(txCtx, payment); err != nil {
				return err
			}

			// sinkronisasi bill
			if status ==
				entity.PaymentStatusCompleted {

				err := s.billSvc.UpdateBillStatus(
					txCtx,
					payment.BillID.String(),
					string(billEntity.BillStatusPaid),
				)

				if err != nil {
					return err
				}
			}

			response = s.entityToResponse(payment)

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *paymentService) entityToResponse(
	payment *entity.Payment,
) *dto.PaymentResponse {

	return &dto.PaymentResponse{
		ID:               payment.ID,
		PaymentNumber:    payment.PaymentNumber,
		BillID:           payment.BillID.String(),
		Amount:           payment.Amount,
		Currency:         payment.Currency,
		Method:           string(payment.Method),
		Status:           payment.Status.String(),
		GatewayReference: payment.GatewayReference,
		PaidAt:           payment.PaidAt,
		Description:      payment.Description,
		CreatedAt:        payment.CreatedAt,
		UpdatedAt:        payment.UpdatedAt,
	}
}

func (s *paymentService) completePayment(payment *entity.Payment) {

	now := time.Now()

	payment.Status = entity.PaymentStatusCompleted
	payment.PaidAt = &now
	payment.UpdatedAt = now
}
