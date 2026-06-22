package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/bill/dto"
	"github.com/novriyantoAli/moodly/internal/application/bill/entity"
	"github.com/novriyantoAli/moodly/internal/application/bill/publisher"
	"github.com/novriyantoAli/moodly/internal/application/bill/repository"

	subscribeDto "github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	subscribeRepository "github.com/novriyantoAli/moodly/internal/application/subscribe/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BillService interface {
	ValidateForPayment(ctx context.Context, billNumber string) (*dto.BillResponse, error)
	LockForPayment(ctx context.Context, billNumber string) (*dto.BillResponse, error)
	CreateBill(ctx context.Context, req *dto.CreateBillRequest) (*dto.BillResponse, error)
	GenerateMonthlyBills(ctx context.Context, month uint, year uint) error
	GetBillByID(ctx context.Context, id string) (*dto.BillResponse, error)
	GetBills(ctx context.Context, filter *dto.BillFilter) (*dto.BillListResponse, error)
	UpdateBillStatus(ctx context.Context, id string, status string) error
	MarkAsPaid(ctx context.Context, id string) error
	SumAmountByMonthYear(ctx context.Context, filter *dto.SumAmountBillFilter) (int64, error)
	CountBillsByMonthYear(ctx context.Context, filter *dto.CountBillFilter) (*dto.CountBillResponse, error)
	ProcessUpdateBillStatusFromUnpaidToOverdue(ctx context.Context, month uint, year uint) error
	QuickCountUnpaidBills(ctx context.Context) (*dto.BillQuickCountUnpaidResponse, error)
	QuickCountOverdueBills(ctx context.Context) (*dto.BillQuickCountOverdueResponse, error)
}

type billService struct {
	client  publisher.BillPublisher
	repo    repository.BillRepository
	subRepo subscribeRepository.SubscribeRepository
	logger  *zap.Logger
}

func NewBillService(repo repository.BillRepository, subRepo subscribeRepository.SubscribeRepository, logger *zap.Logger, client publisher.BillPublisher) BillService {
	return &billService{
		repo:    repo,
		subRepo: subRepo,
		logger:  logger,
		client:  client,
	}
}

func (s *billService) ValidateForPayment(ctx context.Context, billNumber string) (*dto.BillResponse, error) {
	bill, err := s.repo.GetByNumber(ctx, billNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bill not found")
		}
		s.logger.Error("Failed to get bill by bill number", zap.String("bill_number", billNumber), zap.Error(err))
		return nil, err
	}

	if bill.Status != "unpaid" && bill.Status != "overdue" {
		return nil, errors.New("bill is not unpaid or overdue")
	}

	if bill.Amount <= 0 {
		return nil, errors.New("bill value must be greater than 0")
	}

	return s.entityToResponse(bill), nil

}

func (s *billService) LockForPayment(
	ctx context.Context,
	billNumber string,
) (*dto.BillResponse, error) {

	bill, err := s.repo.LockByNumber(
		ctx,
		billNumber,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bill not found")
		}
		s.logger.Error("Failed to lock bill by bill number", zap.String("bill_number", billNumber), zap.Error(err))
		return nil, err
	}

	if bill.Status != "unpaid" &&
		bill.Status != "overdue" {
		return nil, errors.New(
			"bill is not unpaid or overdue",
		)
	}

	if bill.Amount <= 0 {
		return nil, errors.New(
			"bill value must be greater than 0",
		)
	}

	return s.entityToResponse(bill), nil
}

func (s *billService) CreateBill(ctx context.Context, req *dto.CreateBillRequest) (*dto.BillResponse, error) {
	// Validate required fields
	if req.SubscribeID == 0 {
		return nil, errors.New("subscribe_id is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if req.BillMonth < 1 || req.BillMonth > 12 {
		return nil, errors.New("bill_month must be between 1 and 12")
	}

	bill := &entity.Bill{
		ID:          uuid.New(),
		SubscribeID: req.SubscribeID,
		Amount:      req.Amount,
		BillMonth:   req.BillMonth,
		BillYear:    req.BillYear,
		DueDate:     req.DueDate,
		Status:      "unpaid",
		CreatedAt:   time.Now(),
	}

	if req.Status != "" {
		bill.Status = entity.BillStatus(req.Status)
	}

	err := s.repo.Create(ctx, bill)
	if err != nil {
		s.logger.Error("Failed to create bill", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(bill), nil
}

func (s *billService) GenerateMonthlyBills(ctx context.Context, month uint, year uint) error {
	subscribes, err := s.subRepo.GetActiveSubscribes(ctx)
	if err != nil {
		s.logger.Error("Failed to get active subscribes", zap.Error(err))
		return err
	}

	for _, sub := range subscribes {
		// Check if bill already exists for this subscribe, month, and year
		exists, err := s.repo.Exists(ctx, sub.ID, month, year)
		if err != nil {
			s.logger.Info("Failed to check if bill exists", zap.Uint("subscribe_id", sub.ID), zap.Uint("month", month), zap.Uint("year", year), zap.Error(err))
			continue
		}
		if exists {
			s.logger.Info("Bill already exists for this subscribe, month, and year", zap.Uint("subscribe_id", sub.ID), zap.Uint("month", month), zap.Uint("year", year))
			continue
		}

		// Create bill for this subscribe
		req := dto.CreateBillRequest{
			SubscribeID: sub.ID,
			Amount:      int64(sub.Price),
			BillMonth:   month,
			BillYear:    year,
			DueDate:     generateDueDate(sub.StartDate, int(year), time.Month(month)), //time.Date(int(year), time.Month(month), 10, 0, 0, 0, 0, time.Local),
			Status:      "unpaid",
		}

		if req.DueDate.Before(time.Now()) {
			req.Status = "overdue"
		}

		err = s.client.ScheduleBillPerSubscribe(req)
		if err != nil {
			s.logger.Error("Failed to schedule bill generation for subscribe", zap.Uint("subscribe_id", sub.ID), zap.Uint("month", month), zap.Uint("year", year), zap.Error(err))
			continue
		}

		// bill := &entity.Bill{
		// 	ID:          uuid.New(),
		// 	SubscribeID: sub.ID,
		// 	Amount:      int64(sub.Price),
		// 	BillMonth:   month,
		// 	BillYear:    year,
		// 	DueDate:     time.Date(int(year), time.Month(month), 10, 0, 0, 0, 0, time.UTC),
		// 	Status:      "unpaid",
		// 	CreatedAt:   time.Now(),
		// }

		// err = s.repo.Create(ctx, bill)
		// if err != nil {
		// 	s.logger.Error("Failed to create bill", zap.Uint("subscribe_id", sub.ID), zap.Uint("month", month), zap.Uint("year", year), zap.Error(err))
		// 	continue
		// }
		s.logger.Info("Bill generated for subscribe", zap.Uint("subscribe_id", sub.ID), zap.Uint("month", month), zap.Uint("year", year))
	}

	return nil
}

func (s *billService) GetBillByID(ctx context.Context, id string) (*dto.BillResponse, error) {
	// Parse UUID string
	billID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid bill id format")
	}

	bill, err := s.repo.GetByID(ctx, billID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bill not found")
		}
		s.logger.Error("Failed to get bill", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(bill), nil
}

func (s *billService) GetBills(ctx context.Context, filter *dto.BillFilter) (*dto.BillListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	bills, totalCount, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get bills", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.BillResponse, 0, len(bills))
	for _, bill := range bills {
		responses = append(responses, *s.entityToResponse(&bill))
	}

	return &dto.BillListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *billService) MarkAsPaid(ctx context.Context, id string) error {
	// Parse UUID string
	billID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid bill id format")
	}

	// Verify bill exists
	bill, err := s.repo.GetByID(ctx, billID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bill not found")
		}
		return err
	}

	if bill.Status == "paid" {
		return errors.New("bill is already paid")
	}

	err = s.repo.UpdateStatus(ctx, bill.ID, string(entity.BillStatusPaid))
	if err != nil {
		s.logger.Error("Failed to mark bill as paid", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}

func (s *billService) UpdateBillStatus(ctx context.Context, id string, status string) error {
	// Validate status
	if status != "paid" && status != "unpaid" && status != "pending" && status != "overdue" {
		return errors.New("invalid status. allowed values: paid, unpaid, pending, overdue")
	}

	// Parse UUID string
	billID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid bill id format")
	}

	// Verify bill exists
	_, err = s.repo.GetByID(ctx, billID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bill not found")
		}
		return err
	}

	err = s.repo.UpdateStatus(ctx, billID, status)
	if err != nil {
		s.logger.Error("Failed to update bill status", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}

func (s *billService) SumAmountByMonthYear(ctx context.Context, filter *dto.SumAmountBillFilter) (int64, error) {
	if filter.Status != "paid" && filter.Status != "unpaid" && filter.Status != "all" && filter.Status != "overdue" {
		filter.Status = "all"
	}

	if filter.Month < 1 || filter.Month > 12 {
		filter.Month = uint(time.Now().Month())
	}

	if filter.Year <= 0 {
		filter.Year = uint(time.Now().Year())
	}

	amount, err := s.repo.SumAmountByMonthYear(ctx, filter.Status, filter.Month, filter.Year)
	if err != nil {
		s.logger.Error("Failed to sum amount by month and year", zap.Uint("month", filter.Month), zap.Uint("year", filter.Year), zap.Error(err))
		return 0, err
	}

	return amount, nil
}

func (s *billService) CountBillsByMonthYear(ctx context.Context, filter *dto.CountBillFilter) (*dto.CountBillResponse, error) {
	if filter.Status != "paid" && filter.Status != "unpaid" && filter.Status != "all" && filter.Status != "overdue" {
		filter.Status = "all"
	}

	if filter.Month < 1 || filter.Month > 12 {
		filter.Month = uint(time.Now().Month())
	}

	if filter.Year <= 0 {
		filter.Year = uint(time.Now().Year())
	}

	count, err := s.repo.CountBillsByMonthYear(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to count bills by month and year", zap.Uint("month", filter.Month), zap.Uint("year", filter.Year), zap.Error(err))
		return nil, err
	}

	return &dto.CountBillResponse{
		Count:  count,
		Status: filter.Status,
		Month:  filter.Month,
		Year:   filter.Year,
	}, nil
}

func (s *billService) ProcessUpdateBillStatusFromUnpaidToOverdue(ctx context.Context, month uint, year uint) error {
	bills, err := s.repo.GetActiveSubWithUnpaidBills(ctx)
	if err != nil {
		s.logger.Error("Failed to get active subscribes with unpaid bills", zap.Error(err))
		return err
	}

	for _, bill := range bills {
		if bill.DueDate.Before(time.Now()) {
			err = s.client.ScheduleBillPerSubscribeChangeFromUnpaidOverdue(bill)
			if err != nil {
				s.logger.Info("Failed to schedule bill status change from unpaid to overdue", zap.Uint("subscribe_id", bill.SubscribeID), zap.Uint("month", bill.BillMonth), zap.Uint("year", bill.BillYear), zap.Error(err))
				continue
			}
		}
	}
	return nil
}

func (s *billService) QuickCountUnpaidBills(ctx context.Context) (*dto.BillQuickCountUnpaidResponse, error) {
	count, err := s.repo.CountUnpaidBills(ctx)
	if err != nil {
		return nil, err
	}

	amount, err := s.repo.SumUnpaidBillsAmount(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.BillQuickCountUnpaidResponse{Count: count, Amount: amount}, nil
}

func (s *billService) QuickCountOverdueBills(ctx context.Context) (*dto.BillQuickCountOverdueResponse, error) {
	count, err := s.repo.CountOverdueBills(ctx)
	if err != nil {
		return nil, err
	}

	amount, err := s.repo.SumOverdueBillsAmount(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.BillQuickCountOverdueResponse{Count: count, Amount: amount}, nil
}

func generateDueDate(startDate time.Time, year int, month time.Month) time.Time {
	loc := startDate.Location()

	day := startDate.Day()

	// handle bulan yang tidak punya tanggal tsb
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, 23, 59, 59, 0, loc)
}

func (s *billService) entityToResponse(bill *entity.Bill) *dto.BillResponse {
	return &dto.BillResponse{
		ID:          bill.ID,
		SubscribeID: bill.SubscribeID,
		BillNumber:  bill.BillNumber,
		Amount:      bill.Amount,
		BillMonth:   bill.BillMonth,
		BillYear:    bill.BillYear,
		DueDate:     bill.DueDate,
		Status:      string(bill.Status),
		CreatedAt:   bill.CreatedAt,
		UpdatedAt:   bill.UpdatedAt,
		Subscribe: subscribeDto.SubscribeResponse{
			ID:        bill.Subscribe.ID,
			Username:  bill.Subscribe.Username,
			Callname:  bill.Subscribe.CallName,
			Plan:      bill.Subscribe.Plan,
			Price:     bill.Subscribe.Price,
			IsActive:  bill.Subscribe.IsActive,
			CreatedAt: bill.Subscribe.CreatedAt,
			UpdatedAt: bill.Subscribe.UpdatedAt,
		},
	}
}
