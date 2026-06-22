package repository

import (
	"context"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubscribeRepository interface {
	Create(ctx context.Context, subscribe *entity.Subscriber) error
	GetByID(ctx context.Context, id uint) (*entity.Subscriber, error)
	GetByUsername(ctx context.Context, username string) (*entity.Subscriber, error)
	GetActiveSubscribes(ctx context.Context) ([]entity.Subscriber, error)
	GetAll(ctx context.Context, filter *dto.SubscribeFilter) ([]entity.Subscriber, int64, error)
	Update(ctx context.Context, subscribe *entity.Subscriber) error
	Delete(ctx context.Context, id uint) error
	UsernameExists(ctx context.Context, username string) (bool, error)
	Count(ctx context.Context, filter *dto.CountFilter) (int64, error)
}

type subscribeRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewSubscribeRepository(db *gorm.DB, logger *zap.Logger) SubscribeRepository {
	return &subscribeRepository{
		db:     db,
		logger: logger,
	}
}

func (r *subscribeRepository) Create(ctx context.Context, subscribe *entity.Subscriber) error {
	r.logger.Info("Creating subscribe", zap.String("username", subscribe.Username))
	db := database.GetDB(ctx, r.db)
	return db.Create(subscribe).Error
}

func (r *subscribeRepository) GetByID(ctx context.Context, id uint) (*entity.Subscriber, error) {
	var subscribe entity.Subscriber
	db := database.GetDB(ctx, r.db)
	err := db.First(&subscribe, id).Error
	if err != nil {
		r.logger.Error("Failed to get subscribe by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &subscribe, nil
}

func (r *subscribeRepository) GetByUsername(ctx context.Context, username string) (*entity.Subscriber, error) {
	var subscribe entity.Subscriber
	db := database.GetDB(ctx, r.db)
	err := db.Where("username = ?", username).First(&subscribe).Error
	if err != nil {
		r.logger.Error("Failed to get subscribe by username", zap.String("username", username), zap.Error(err))
		return nil, err
	}
	return &subscribe, nil
}

func (r *subscribeRepository) GetActiveSubscribes(ctx context.Context) ([]entity.Subscriber, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	var subscribes []entity.Subscriber
	db := database.GetDB(ctx, r.db)

	err := db.Where("is_active = ?", true).
		Where(`
    NOT EXISTS (
        SELECT 1 FROM bills b
        WHERE b.subscribe_id = subscribers.id
        AND b.bill_month = ?
        AND b.bill_year = ?
    )`, month, year).Find(&subscribes).Error

	// err := db.Where("is_active = ?", true).Find(&subscribes).Error
	if err != nil {
		r.logger.Error("Failed to get active subscribes", zap.Error(err))
		return nil, err
	}
	return subscribes, nil
}

func (r *subscribeRepository) GetAll(ctx context.Context, filter *dto.SubscribeFilter) ([]entity.Subscriber, int64, error) {
	db := database.GetDB(ctx, r.db)
	var subscribes []entity.Subscriber
	var totalCount int64

	query := db.Model(&entity.Subscriber{})

	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}

	if filter.CallName != "" {
		query = query.Where("call_name LIKE ?", "%"+filter.CallName+"%")
	}

	err := query.Count(&totalCount).Error
	if err != nil {
		r.logger.Error("Failed to count subscribes", zap.Error(err))
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err = query.Find(&subscribes).Error
	if err != nil {
		r.logger.Error("Failed to get subscribes", zap.Error(err))
		return nil, 0, err
	}

	return subscribes, totalCount, nil
}

func (r *subscribeRepository) Update(ctx context.Context, subscribe *entity.Subscriber) error {
	r.logger.Info("Updating subscribe", zap.Uint("id", subscribe.ID))
	db := database.GetDB(ctx, r.db)
	return db.Save(subscribe).Error
}

func (r *subscribeRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting subscribe", zap.Uint("id", id))
	db := database.GetDB(ctx, r.db)
	return db.Delete(&entity.Subscriber{}, id).Error
}

func (r *subscribeRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var count int64
	db := database.GetDB(ctx, r.db)
	err := db.Model(&entity.Subscriber{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to check if username exists", zap.String("username", username), zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

func (r *subscribeRepository) Count(ctx context.Context, filter *dto.CountFilter) (int64, error) {
	db := database.GetDB(ctx, r.db)
	var count int64

	query := db.Model(&entity.Subscriber{})

	if filter.Plan != "" {
		query = query.Where("plan = ?", filter.Plan)
	}

	query = query.Where("is_active = ?", filter.IsActive)

	err := query.Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to count subscribes", zap.Error(err))
		return 0, err
	}

	return count, nil
}
