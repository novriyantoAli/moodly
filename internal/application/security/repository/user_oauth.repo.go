package repository

import (
	"context"

	"github.com/novriyantoAli/moodly/internal/application/security/entity"
	database "github.com/novriyantoAli/moodly/internal/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserOAuthRepository interface {
	Create(ctx context.Context, oauth *entity.UserOAuth) error

	GetByProviderAndUserID(
		ctx context.Context,
		provider string,
		providerUserID string,
	) (*entity.UserOAuth, error)

	GetByUserID(
		ctx context.Context,
		userID uint,
	) ([]*entity.UserOAuth, error)

	Delete(
		ctx context.Context,
		userID uint,
		provider string,
	) error
}

type userOAuthRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserOAuthRepository(db *gorm.DB, logger *zap.Logger) UserOAuthRepository {
	return &userOAuthRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userOAuthRepository) Create(ctx context.Context, oauth *entity.UserOAuth) error {

	r.logger.Info(
		"Creating user OAuth",
		zap.Uint("user_id", oauth.UserID),
	)

	db := database.GetDB(ctx, r.db)

	return db.Create(oauth).Error
}

func (r *userOAuthRepository) GetByProviderAndUserID(
	ctx context.Context,
	provider string,
	providerUserID string,
) (*entity.UserOAuth, error) {

	var record entity.UserOAuth

	db := database.GetDB(ctx, r.db)
	err := db.
		WithContext(ctx).
		Where(
			"provider = ? AND provider_user_id = ?",
			provider,
			providerUserID,
		).
		First(&record).
		Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userOAuthRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) ([]*entity.UserOAuth, error) {

	var records []entity.UserOAuth

	db := database.GetDB(ctx, r.db)
	err := db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&records).
		Error

	if err != nil {
		return nil, err
	}

	result := make([]*entity.UserOAuth, 0, len(records))

	for _, record := range records {
		result = append(result, &entity.UserOAuth{
			ID:             record.ID,
			UserID:         record.UserID,
			Provider:       record.Provider,
			ProviderUserID: record.ProviderUserID,
			Email:          record.Email,
			Name:           record.Name,
			Picture:        record.Picture,
			CreatedAt:      record.CreatedAt,
		})
	}

	return result, nil
}

func (r *userOAuthRepository) Delete(
	ctx context.Context,
	userID uint,
	provider string,
) error {

	db := database.GetDB(ctx, r.db)
	return db.
		WithContext(ctx).
		Where(
			"user_id = ? AND provider = ?",
			userID,
			provider,
		).
		Delete(&entity.UserOAuth{}).
		Error
}
