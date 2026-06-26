package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/authorization/entity"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthorizationRepository interface {
	GetRolesByUserID(ctx context.Context, userID uint) ([]entity.Role, error)
	GetPermissionsByRoles(ctx context.Context, roleNames []string) ([]string, error)
	InvalidateUserCache(ctx context.Context, userID uint) error
}

type authorizationRepository struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

func NewAuthorizationRepository(db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) AuthorizationRepository {
	return &authorizationRepository{
		db:     db,
		redis:  redisClient,
		logger: logger,
	}
}

func (r *authorizationRepository) GetRolesByUserID(ctx context.Context, userID uint) ([]entity.Role, error) {
	cacheKey := fmt.Sprintf("auth:user:%d:roles", userID)
	
	// Try getting from cache
	if r.redis != nil {
		cachedRoles, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var roles []entity.Role
			if err := json.Unmarshal([]byte(cachedRoles), &roles); err == nil {
				return roles, nil
			}
		}
	}

	var roles []entity.Role
	err := r.db.WithContext(ctx).
		Select("roles.*").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	// Set cache
	if r.redis != nil && len(roles) > 0 {
		if rolesBytes, err := json.Marshal(roles); err == nil {
			r.redis.Set(ctx, cacheKey, rolesBytes, 15*time.Minute)
		}
	}

	return roles, nil
}

func (r *authorizationRepository) GetPermissionsByRoles(ctx context.Context, roleNames []string) ([]string, error) {
	if len(roleNames) == 0 {
		return []string{}, nil
	}

	cacheKey := fmt.Sprintf("auth:roles:%v:permissions", roleNames)

	// Try getting from cache
	if r.redis != nil {
		cachedPerms, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var permissions []string
			if err := json.Unmarshal([]byte(cachedPerms), &permissions); err == nil {
				return permissions, nil
			}
		}
	}

	var permissions []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.code").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Where("roles.name IN ?", roleNames).
		Pluck("code", &permissions).Error

	if err != nil {
		return nil, err
	}

	// Set cache
	if r.redis != nil && len(permissions) > 0 {
		if permsBytes, err := json.Marshal(permissions); err == nil {
			r.redis.Set(ctx, cacheKey, permsBytes, 15*time.Minute)
		}
	}

	return permissions, nil
}

func (r *authorizationRepository) InvalidateUserCache(ctx context.Context, userID uint) error {
	if r.redis != nil {
		cacheKey := fmt.Sprintf("auth:user:%d:roles", userID)
		return r.redis.Del(ctx, cacheKey).Err()
	}
	return nil
}
