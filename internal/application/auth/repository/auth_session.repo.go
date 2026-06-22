package repository

import (
	"context"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/auth/entity"
	database "github.com/novriyantoAli/moodly/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthSessionRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.AuthSession, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.AuthSession, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.AuthSession, error)
	Create(ctx context.Context, session *entity.AuthSession) error
	UpdateAccessToken(ctx context.Context, id uint, accessToken string) error
	UpdateRefreshToken(ctx context.Context, id uint, refreshToken string, expiredAt time.Time) error
	Delete(ctx context.Context, id uint) error
	DeleteByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteByUserID(ctx context.Context, userID uint) error
	DeleteExpiredSessions(ctx context.Context) error
	GetActiveSessionCount(ctx context.Context, userID uint) (int64, error)
	GetOldestActiveSession(ctx context.Context, userID uint) (*entity.AuthSession, error)
}

type authSessionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewAuthSessionRepository(
	db *gorm.DB,
	logger *zap.Logger,
) AuthSessionRepository {

	return &authSessionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *authSessionRepository) GetByID(ctx context.Context, id uint) (*entity.AuthSession, error) {

	r.logger.Info("Getting session by id", zap.Uint("id", id))

	db := database.GetDB(ctx, r.db)

	var session entity.AuthSession

	if err := db.
		Where("id = ?", id).
		First(&session).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &session, nil
}

func (r *authSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.AuthSession, error) {
	r.logger.Info("Getting session by refresh token")

	db := database.GetDB(ctx, r.db)

	var session entity.AuthSession

	if err := db.Where("refresh_token = ?", refreshToken).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}

func (r *authSessionRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.AuthSession, error) {

	r.logger.Info(
		"Getting sessions by user",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	var sessions []entity.AuthSession

	if err := db.
		Where("user_id = ?", userID).
		Find(&sessions).Error; err != nil {

		return nil, err
	}

	return sessions, nil
}

func (r *authSessionRepository) Create(ctx context.Context, session *entity.AuthSession) error {
	r.logger.Info(
		"Creating authentication session",
		zap.Uint("user_id", session.UserID),
	)

	db := database.GetDB(ctx, r.db)

	return db.Create(session).Error
}

func (r *authSessionRepository) UpdateAccessToken(
	ctx context.Context,
	id uint,
	accessToken string,
) error {

	r.logger.Info(
		"Updating access token",
		zap.Uint("id", id),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Model(&entity.AuthSession{}).
		Where("id = ?", id).
		Update("access_token", accessToken).
		Error
}

func (r *authSessionRepository) UpdateRefreshToken(
	ctx context.Context,
	id uint,
	refreshToken string,
	expiredAt time.Time,
) error {

	r.logger.Info(
		"Updating refresh token",
		zap.Uint("id", id),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Model(&entity.AuthSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"refresh_token": refreshToken,
			"expired_at":    expiredAt,
		}).Error
}

func (r *authSessionRepository) Delete(
	ctx context.Context,
	id uint,
) error {

	r.logger.Info(
		"Deleting session",
		zap.Uint("id", id),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Where("id = ?", id).
		Delete(&entity.AuthSession{}).
		Error
}

func (r *authSessionRepository) DeleteByRefreshToken(
	ctx context.Context,
	refreshToken string,
) error {

	r.logger.Info(
		"Deleting session by refresh token",
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Where("refresh_token = ?", refreshToken).
		Delete(&entity.AuthSession{}).
		Error
}

func (r *authSessionRepository) DeleteByUserID(
	ctx context.Context,
	userID uint,
) error {

	r.logger.Info(
		"Deleting all user sessions",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Where("user_id = ?", userID).
		Delete(&entity.AuthSession{}).
		Error
}

func (r *authSessionRepository) DeleteExpiredSessions(ctx context.Context) error {

	r.logger.Info(
		"Deleting expired sessions",
	)

	db := database.GetDB(ctx, r.db)

	return db.
		Where("expired_at < ?", time.Now().UTC()).
		Delete(&entity.AuthSession{}).
		Error
}

func (r *authSessionRepository) GetActiveSessionCount(
	ctx context.Context,
	userID uint,
) (int64, error) {

	r.logger.Info(
		"Getting active session count",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	var count int64

	err := db.
		Model(&entity.AuthSession{}).
		Where(
			"user_id = ? AND expired_at > ?",
			userID,
			time.Now().UTC(),
		).
		Count(&count).
		Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *authSessionRepository) GetOldestActiveSession(ctx context.Context, userID uint) (*entity.AuthSession, error) {

	r.logger.Info(
		"Getting oldest active session",
		zap.Uint("user_id", userID),
	)

	db := database.GetDB(ctx, r.db)

	var session entity.AuthSession

	err := db.
		Where(
			"user_id = ? AND expired_at > ?",
			userID,
			time.Now(),
		).
		Order("created_at ASC").
		First(&session).
		Error

	if err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &session, nil
}
