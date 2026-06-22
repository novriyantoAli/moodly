package repository

import (
	"context"

	"github.com/novriyantoAli/moodly/internal/application/auth/entity"
	database "github.com/novriyantoAli/moodly/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type LoginAttemptRepository interface {
	Create(ctx context.Context, attempt *entity.LoginAttempt) error
	GetByUserID(ctx context.Context, userID uint) ([]entity.LoginAttempt, error)
	GetByUsername(ctx context.Context, username string) ([]entity.LoginAttempt, error)
}

type loginAttemptRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewLoginAttemptRepository(db *gorm.DB, logger *zap.Logger) LoginAttemptRepository {
	return &loginAttemptRepository{
		db:     db,
		logger: logger,
	}
}

func (r *loginAttemptRepository) Create(ctx context.Context, attempt *entity.LoginAttempt) error {
	r.logger.Info(
		"Creating login attempt",
		zap.String("username", attempt.Username),
	)

	db := database.GetDB(ctx, r.db)

	return db.Create(attempt).Error
}

func (r *loginAttemptRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) ([]entity.LoginAttempt, error) {

	db := database.GetDB(ctx, r.db)

	var attempts []entity.LoginAttempt

	if err := db.
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&attempts).Error; err != nil {

		return nil, err
	}

	return attempts, nil
}

func (r *loginAttemptRepository) GetByUsername(
	ctx context.Context,
	username string,
) ([]entity.LoginAttempt, error) {

	db := database.GetDB(ctx, r.db)

	var attempts []entity.LoginAttempt

	if err := db.
		Where("username = ?", username).
		Order("created_at desc").
		Find(&attempts).Error; err != nil {

		return nil, err
	}

	return attempts, nil
}
