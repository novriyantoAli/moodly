package service

import (
	"context"

	"github.com/novriyantoAli/moodly/internal/application/authorization/entity"
	"github.com/novriyantoAli/moodly/internal/application/authorization/repository"
	"go.uber.org/zap"
)

type AuthorizationService interface {
	GetRolesByUserID(ctx context.Context, userID uint) ([]entity.Role, error)
	GetPermissionsByRoles(ctx context.Context, roles []string) ([]string, error)
	IsPsychologist(ctx context.Context, userID uint) (bool, error)
}

type authorizationService struct {
	authRepo repository.AuthorizationRepository
	logger   *zap.Logger
}

func NewAuthorizationService(authRepo repository.AuthorizationRepository, logger *zap.Logger) AuthorizationService {
	return &authorizationService{
		authRepo: authRepo,
		logger:   logger,
	}
}

func (s *authorizationService) GetRolesByUserID(ctx context.Context, userID uint) ([]entity.Role, error) {
	return s.authRepo.GetRolesByUserID(ctx, userID)
}

func (s *authorizationService) GetPermissionsByRoles(ctx context.Context, roles []string) ([]string, error) {
	return s.authRepo.GetPermissionsByRoles(ctx, roles)
}

func (s *authorizationService) IsPsychologist(ctx context.Context, userID uint) (bool, error) {
	roles, err := s.GetRolesByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == "psikologi" || role.Name == "psikolog" {
			return true, nil
		}
	}

	return false, nil
}
