package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/bill/dto"
	"github.com/novriyantoAli/moodly/internal/application/bill/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BillRepository interface {
	Create(ctx context.Context, bill *entity.Bill) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Bill, error)
	GetByNumber(ctx context.Context, billNumber string) (*entity.Bill, error)
	LockByNumber(ctx context.Context, billNumber string) (*entity.Bill, error)
	GetAll(ctx context.Context, filter *dto.BillFilter) ([]entity.Bill, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Exists(ctx context.Context, subscribeID uint, month uint, year uint) (bool, error)
	GetActiveSubWithUnpaidBills(ctx context.Context) ([]entity.Bill, error)
	CountUnpaidBills(ctx context.Context) (int64, error)
	SumUnpaidBillsAmount(ctx context.Context) (int64, error)
	SumAmountByMonthYear(ctx context.Context, status string, month uint, year uint) (int64, error)
	CountBillsByMonthYear(ctx context.Context, filter *dto.CountBillFilter) (int64, error)
	CountOverdueBills(ctx context.Context) (int64, error)
	SumOverdueBillsAmount(ctx context.Context) (int64, error)
}

type billRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewBillRepository(db *gorm.DB, logger *zap.Logger) BillRepository {
	return &billRepository{
		db:     db,
		logger: logger,
	}
}

func (r *billRepository) Create(ctx context.Context, bill *entity.Bill) error {
	r.logger.Info("Creating bill", zap.Uint("subscribe_id", bill.SubscribeID))
	db := database.GetDB(ctx, r.db)
	return db.Create(bill).Error
}

func (r *billRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Bill, error) {
	var bill entity.Bill
	db := database.GetDB(ctx, r.db)
	err := db.First(&bill, "id = ?", id).Error
	if err != nil {
		r.logger.Error("Failed to get bill by ID", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	return &bill, nil
}

func (r *billRepository) GetByNumber(ctx context.Context, billNumber string) (*entity.Bill, error) {
	var bill entity.Bill
	db := database.GetDB(ctx, r.db)
	err := db.First(&bill, "bill_number = ?", billNumber).Error
	if err != nil {
		r.logger.Error("Failed to get bill by number", zap.String("bill_number", billNumber), zap.Error(err))
		return nil, err
	}
	return &bill, nil
}

func (r *billRepository) LockByNumber(
	ctx context.Context,
	billNumber string,
) (*entity.Bill, error) {

	db := database.GetDB(
		ctx,
		r.db,
	)

	var bill entity.Bill

	err := db.
		Clauses(
			clause.Locking{
				Strength: "UPDATE",
			},
		).
		Where(
			"bill_number = ?",
			billNumber,
		).
		First(&bill).
		Error

	if err != nil {
		return nil, err
	}

	return &bill, nil
}

func (r *billRepository) GetAll(ctx context.Context, filter *dto.BillFilter) ([]entity.Bill, int64, error) {
	var bills []entity.Bill
	var totalCount int64

	db := database.GetDB(ctx, r.db)
	query := db.Model(&entity.Bill{})

	if filter.SubscribeID > 0 {
		query = query.Where("subscribe_id = ?", filter.SubscribeID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	query.Count(&totalCount)

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Preload("Subscribe").Find(&bills).Error
	if err != nil {
		r.logger.Error("Failed to get bills", zap.Error(err))
		return nil, 0, err
	}

	return bills, totalCount, nil
}

func (r *billRepository) GetActiveSubWithUnpaidBills(ctx context.Context) ([]entity.Bill, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	var bills []entity.Bill
	db := database.GetDB(ctx, r.db)

	err := db.Joins("JOIN subscribers ON subscribers.id = bills.subscribe_id").
		Where("subscribers.is_active = ?", true).
		Where("bills.status = ?", "unpaid").
		Where("bills.bill_month = ?", month).
		Where("bills.bill_year = ?", year).
		Preload("Subscribe").
		Find(&bills).Error

	if err != nil {
		r.logger.Error("Failed to get active subscribes with unpaid bills", zap.Error(err))
		return nil, err
	}

	return bills, nil
}
func (r *billRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	r.logger.Info("Updating bill status", zap.String("bill_id", id.String()), zap.String("status", status))
	db := database.GetDB(ctx, r.db)
	err := db.Model(&entity.Bill{}).Where("id = ?", id).Update("status", status).Error
	if err != nil {
		r.logger.Error("Failed to update bill status", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

func (r *billRepository) Exists(ctx context.Context, subscribeID uint, month uint, year uint) (bool, error) {
	var count int64
	db := database.GetDB(ctx, r.db)
	err := db.Model(&entity.Bill{}).Where("subscribe_id = ? AND bill_month = ? AND bill_year = ?", subscribeID, month, year).Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to check if bill exists", zap.Uint("subscribe_id", subscribeID), zap.Uint("month", month), zap.Uint("year", year), zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

func (r *billRepository) SumAmountByMonthYear(ctx context.Context, status string, month uint, year uint) (int64, error) {
	var totalAmount int64
	db := database.GetDB(ctx, r.db)

	if status == "all" {
		err := db.Model(&entity.Bill{}).Select("COALESCE(SUM(amount), 0)").Where("bill_month = ? AND bill_year = ?", month, year).Scan(&totalAmount).Error
		if err != nil {
			r.logger.Error("Failed to sum bills amount by month and year", zap.Uint("month", month), zap.Uint("year", year), zap.Error(err))
			return 0, err
		}
	} else {
		err := db.Model(&entity.Bill{}).Select("COALESCE(SUM(amount), 0)").Where("status = ? AND bill_month = ? AND bill_year = ?", status, month, year).Scan(&totalAmount).Error
		if err != nil {
			r.logger.Error("Failed to sum bills amount by month and year", zap.String("status", status), zap.Uint("month", month), zap.Uint("year", year), zap.Error(err))
			return 0, err
		}
	}
	return totalAmount, nil
}

func (r *billRepository) CountBillsByMonthYear(ctx context.Context, filter *dto.CountBillFilter) (int64, error) {
	var count int64
	db := database.GetDB(ctx, r.db)

	query := db.Model(&entity.Bill{}).Where("bill_month = ? AND bill_year = ?", filter.Month, filter.Year)

	if filter.Status != "all" {
		query = query.Where("status = ?", filter.Status)
	}

	err := query.Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to count bills by month and year", zap.Uint("month", filter.Month), zap.Uint("year", filter.Year), zap.Error(err))
		return 0, err
	}

	return count, nil
}

func (r *billRepository) CountUnpaidBills(ctx context.Context) (int64, error) {
	var count int64
	db := database.GetDB(ctx, r.db)
	err := db.Model(&entity.Bill{}).Where("status = ?", "unpaid").Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to count unpaid bills", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *billRepository) SumUnpaidBillsAmount(ctx context.Context) (int64, error) {
	var totalAmount int64
	db := database.GetDB(ctx, r.db)

	err := db.Model(&entity.Bill{}).Select("COALESCE(SUM(amount), 0)").Where("status = ?", "unpaid").Scan(&totalAmount).Error
	if err != nil {
		r.logger.Error("Failed to sum unpaid bills amount", zap.Error(err))
		return 0, err
	}
	return totalAmount, nil
}

func (r *billRepository) CountOverdueBills(ctx context.Context) (int64, error) {
	var count int64
	db := database.GetDB(ctx, r.db)
	err := db.Model(&entity.Bill{}).Where("status = ?", "overdue").Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to count overdue bills", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *billRepository) SumOverdueBillsAmount(ctx context.Context) (int64, error) {
	var totalAmount int64
	db := database.GetDB(ctx, r.db)

	err := db.Model(&entity.Bill{}).Select("COALESCE(SUM(amount), 0)").Where("status = ?", "overdue").Scan(&totalAmount).Error
	if err != nil {
		r.logger.Error("Failed to sum overdue bills amount", zap.Error(err))
		return 0, err
	}
	return totalAmount, nil
}
