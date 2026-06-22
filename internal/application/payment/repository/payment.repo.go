package repository

import (
	"context"

	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	GetByID(ctx context.Context, id uint) (*entity.Payment, error)
	GetByIDForUpdate(ctx context.Context, id uint) (*entity.Payment, error)
	GetByPaymentNumber(ctx context.Context, paymentNumber string) (*entity.Payment, error)
	GetByBillID(ctx context.Context, billID string) ([]entity.Payment, error)
	GetAll(ctx context.Context, filter *dto.PaymentFilter) ([]entity.Payment, int64, error)
	Update(ctx context.Context, payment *entity.Payment) error
	ExistsActivePaymentByBillID(ctx context.Context, billID string) (bool, error)
}

type paymentRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPaymentRepository(db *gorm.DB, logger *zap.Logger) PaymentRepository {
	return &paymentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *paymentRepository) Create(
	ctx context.Context,
	payment *entity.Payment,
) error {

	db := database.GetDB(ctx, r.db)

	r.logger.Info(
		"Creating payment",
		zap.String(
			"payment_number",
			payment.PaymentNumber,
		),
		zap.String(
			"bill_id",
			payment.BillID.String(),
		),
	)

	return db.Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx context.Context, id uint) (*entity.Payment, error) {

	db := database.GetDB(ctx, r.db)

	var payment entity.Payment

	err := db.
		Preload("Bill").
		First(&payment, id).
		Error

	if err != nil {

		r.logger.Error(
			"Failed to get payment by ID",
			zap.Uint("id", id),
			zap.Error(err),
		)

		return nil, err
	}

	return &payment, nil
}

func (r *paymentRepository) GetByIDForUpdate(
	ctx context.Context,
	id uint,
) (*entity.Payment, error) {

	db := database.GetDB(ctx, r.db)

	var payment entity.Payment

	err := db.
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).
		First(&payment, id).
		Error

	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *paymentRepository) GetByPaymentNumber(ctx context.Context, paymentNumber string) (*entity.Payment, error) {

	db := database.GetDB(ctx, r.db)

	var payment entity.Payment

	err := db.
		Preload("Bill").
		Where(
			"payment_number = ?",
			paymentNumber,
		).
		First(&payment).
		Error

	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *paymentRepository) GetByBillID(
	ctx context.Context,
	billID string,
) ([]entity.Payment, error) {

	db := database.GetDB(ctx, r.db)

	var payments []entity.Payment

	err := db.
		Preload("Bill").
		Where(
			"bill_id = ?",
			billID,
		).
		Find(&payments).
		Error

	if err != nil {

		r.logger.Error(
			"Failed to get payments by bill ID",
			zap.String("bill_id", billID),
			zap.Error(err),
		)

		return nil, err
	}

	return payments, nil
}

func (r *paymentRepository) GetAll(ctx context.Context, filter *dto.PaymentFilter) ([]entity.Payment, int64, error) {
	db := database.GetDB(ctx, r.db)
	var payments []entity.Payment
	var totalCount int64

	query := db.Model(&entity.Payment{})

	if filter.Status != "" {
		query = query.Where(
			"status = ?",
			filter.Status,
		)
	}

	if filter.BillID != "" {
		query = query.Where(
			"bill_id = ?",
			filter.BillID,
		)
	}

	if filter.PaymentNumber != "" {
		query = query.Where(
			"payment_number = ?",
			filter.PaymentNumber,
		)
	}

	if filter.Method != "" {
		query = query.Where(
			"method = ?",
			filter.Method,
		)
	}

	if filter.Currency != "" {
		query = query.Where(
			"currency = ?",
			filter.Currency,
		)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		r.logger.Error("Failed to count payments", zap.Error(err))
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Preload("Bill").Find(&payments).Error
	if err != nil {
		r.logger.Error("Failed to get payments", zap.Error(err))
		return nil, 0, err
	}

	return payments, totalCount, nil
}

func (r *paymentRepository) Update(
	ctx context.Context,
	payment *entity.Payment,
) error {

	db := database.GetDB(ctx, r.db)

	r.logger.Info(
		"Updating payment",
		zap.Uint("id", payment.ID),
	)

	// update hanya field yang diijinkan
	// seperti status, method, gateway reference, description, paid at

	return db.Model(&entity.Payment{}).
		Where("id = ?", payment.ID).
		Updates(map[string]interface{}{
			"status":      payment.Status,
			"description": payment.Description,
			"paid_at":     payment.PaidAt,
		}).Error
}

func (r *paymentRepository) ExistsActivePaymentByBillID(ctx context.Context, billID string) (bool, error) {

	db := database.GetDB(ctx, r.db)

	var count int64

	err := db.
		Model(&entity.Payment{}).
		Where(
			"bill_id = ? AND status IN ?",
			billID,
			[]string{
				entity.PaymentStatusPending.String(),
				entity.PaymentStatusCompleted.String(),
			},
		).
		Count(&count).
		Error

	if err != nil {
		r.logger.Error(
			"Failed to check active payment by bill id",
			zap.String("bill_id", billID),
			zap.Error(err),
		)
		return false, err
	}

	return count > 0, nil
}
