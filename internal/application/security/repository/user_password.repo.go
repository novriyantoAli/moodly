package repository

import (
	"context"
	"time"

	database "github.com/novriyantoAli/moodly/internal/pkg/database"

	"github.com/novriyantoAli/moodly/internal/application/security/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserPasswordRepository interface {
	GetByUserID(ctx context.Context, userID uint) (*entity.UserPassword, error)
	GetByUsername(ctx context.Context, username string) (*entity.UserPassword, error)
	Create(ctx context.Context, password *entity.UserPassword) error
	UpdatePasswordHash(
		ctx context.Context,
		userID uint,
		passwordHash string,
	) error
	IncrementFailedAttempt(
		ctx context.Context,
		userID uint,
	) error
	ResetFailedAttempt(
		ctx context.Context,
		userID uint,
	) error
	LockAccount(
		ctx context.Context,
		userID uint,
		duration time.Duration,
	) error
	Unlock(
		ctx context.Context,
		userID uint,
	) error
	UpdateLastLogin(
		ctx context.Context,
		userID uint,
		loginTime time.Time,
	) error
	Delete(
		ctx context.Context,
		userID uint,
	) error
}

type userPasswordRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserPasswordRepository(
	db *gorm.DB,
	logger *zap.Logger,
) UserPasswordRepository {
	return &userPasswordRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userPasswordRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) (*entity.UserPassword, error) {

	r.logger.Info(
		"Getting user password",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	var password entity.UserPassword

	if err := db.
		Where("user_id = ?", userID).
		First(&password).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &password, nil
}

func (r *userPasswordRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*entity.UserPassword, error) {

	r.logger.Info(
		"Getting user password by username",
		zap.String("username", username),
	)

	db := database.GetDB(ctx, r.db)

	var password entity.UserPassword

	if err := db.
		Where("username = ?", username).
		First(&password).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &password, nil
}

func (r *userPasswordRepository) Create(
	ctx context.Context,
	password *entity.UserPassword,
) error {

	r.logger.Info(
		"Creating user password",
		zap.Uint("user_id", password.UserID),
	)

	db := database.GetDB(ctx, r.db)

	return db.Create(password).Error
}

func (r *userPasswordRepository) UpdatePasswordHash(
	ctx context.Context,
	userID uint,
	passwordHash string,
) error {

	r.logger.Info(
		"Updating password hash",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Model(&entity.UserPassword{}).
		Where("user_id = ?", userID).
		Update("password_hash", passwordHash).
		Error
}

func (r *userPasswordRepository) IncrementFailedAttempt(
	ctx context.Context,
	userID uint,
) error {

	r.logger.Info(
		"Increment failed attempt",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Model(&entity.UserPassword{}).
		Where("user_id = ?", userID).
		Update(
			"failed_attempt",
			gorm.Expr("failed_attempt + 1"),
		).Error
}

func (r *userPasswordRepository) LockAccount(
	ctx context.Context,
	userID uint,
	duration time.Duration,
) error {

	r.logger.Info(
		"Lock account",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	lockedUntil := time.Now().Add(duration)

	return db.
		Model(&entity.UserPassword{}).
		Where("user_id = ?", userID).
		Update("locked_until", lockedUntil).
		Error
}

func (r *userPasswordRepository) ResetFailedAttempt(
	ctx context.Context,
	userID uint,
) error {

	r.logger.Info(
		"Reset failed attempt",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Model(&entity.UserPassword{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"failed_attempt": 0,
			"locked_until":   nil,
		}).Error
}

func (r *userPasswordRepository) Unlock(
	ctx context.Context,
	userID uint,
) error {

	r.logger.Info(
		"Unlock account",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Model(&entity.UserPassword{}).
		Where("user_id = ?", userID).
		Update("locked_until", nil).
		Error
}

func (r *userPasswordRepository) UpdateLastLogin(
	ctx context.Context,
	userID uint,
	loginTime time.Time,
) error {

	r.logger.Info(
		"Updating last login",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Model(&entity.UserPassword{}).
		Where("user_id = ?", userID).
		Update("last_login_at", loginTime).
		Error
}

func (r *userPasswordRepository) Delete(
	ctx context.Context,
	userID uint,
) error {

	r.logger.Info(
		"Deleting user password",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Where("user_id = ?", userID).
		Delete(&entity.UserPassword{}).
		Error
}
