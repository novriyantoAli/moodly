package service

import (
	"context"
	"fmt"
	"time"

	entity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	repository "github.com/novriyantoAli/moodly/internal/application/auth/repository"

	"go.uber.org/zap"
)

const (
	MaxActiveSessions = 2
)

type AuthSessionService interface {
	CreateSession(ctx context.Context, session *entity.AuthSession) error
	GetSessionByID(ctx context.Context, id uint) (*entity.AuthSession, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*entity.AuthSession, error)
	GetUserSessions(ctx context.Context, userID uint) ([]entity.AuthSession, error)
	RefreshSession(ctx context.Context, refreshToken string, newAccessToken string, newRefreshToken string, expiredAt time.Time) error
	Logout(ctx context.Context, sessionID uint) error
	LogoutByRefreshToken(ctx context.Context, refreshToken string) error
	LogoutAllUserSessions(ctx context.Context, userID uint) error
	DeleteExpiredSessions(ctx context.Context) error
	GetActiveSessionCount(ctx context.Context, userID uint) (int64, error)
}

type authSessionService struct {
	repo   repository.AuthSessionRepository
	logger *zap.Logger
}

func NewAuthSessionService(
	repo repository.AuthSessionRepository,
	logger *zap.Logger,
) AuthSessionService {

	return &authSessionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *authSessionService) CreateSession(
	ctx context.Context,
	session *entity.AuthSession,
) error {

	s.logger.Info(
		"Creating auth session",
		zap.Uint("user_id", session.UserID),
	)

	activeCount, err := s.repo.GetActiveSessionCount(
		ctx,
		session.UserID,
	)
	if err != nil {
		return err
	}

	if activeCount >= MaxActiveSessions {

		oldestSession, err := s.repo.GetOldestActiveSession(
			ctx,
			session.UserID,
		)
		if err != nil {
			return err
		}

		if oldestSession != nil {
			if err := s.repo.Delete(
				ctx,
				oldestSession.ID,
			); err != nil {
				return err
			}
		}
	}

	return s.repo.Create(
		ctx,
		session,
	)
}

func (s *authSessionService) GetSessionByID(
	ctx context.Context,
	id uint,
) (*entity.AuthSession, error) {

	return s.repo.GetByID(
		ctx,
		id,
	)
}

func (s *authSessionService) GetSessionByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (*entity.AuthSession, error) {

	session, err := s.repo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, nil
	}

	if session.ExpiredAt.Before(time.Now().UTC()) {
		return nil, nil
	}

	return session, nil
}

func (s *authSessionService) GetUserSessions(
	ctx context.Context,
	userID uint,
) ([]entity.AuthSession, error) {

	return s.repo.GetByUserID(
		ctx,
		userID,
	)
}

func (s *authSessionService) RefreshSession(
	ctx context.Context,
	refreshToken string,
	newAccessToken string,
	newRefreshToken string,
	expiredAt time.Time,
) error {

	session, err := s.repo.GetByRefreshToken(
		ctx,
		refreshToken,
	)
	if err != nil {
		return err
	}

	if session == nil {
		return fmt.Errorf("session not found")
	}

	if session.ExpiredAt.Before(time.Now().UTC()) {
		return fmt.Errorf("refresh token expired")
	}

	if err := s.repo.UpdateAccessToken(
		ctx,
		session.ID,
		newAccessToken,
	); err != nil {
		return err
	}

	return s.repo.UpdateRefreshToken(
		ctx,
		session.ID,
		newRefreshToken,
		expiredAt,
	)
}

func (s *authSessionService) Logout(
	ctx context.Context,
	sessionID uint,
) error {

	s.logger.Info(
		"Logout session",
		zap.Uint("session_id", sessionID),
	)

	return s.repo.Delete(
		ctx,
		sessionID,
	)
}

func (s *authSessionService) LogoutByRefreshToken(
	ctx context.Context,
	refreshToken string,
) error {

	s.logger.Info(
		"Logout by refresh token",
	)

	return s.repo.DeleteByRefreshToken(
		ctx,
		refreshToken,
	)
}

func (s *authSessionService) LogoutAllUserSessions(
	ctx context.Context,
	userID uint,
) error {

	s.logger.Info(
		"Logout all user sessions",
		zap.Uint("user_id", userID),
	)

	return s.repo.DeleteByUserID(
		ctx,
		userID,
	)
}

func (s *authSessionService) DeleteExpiredSessions(
	ctx context.Context,
) error {

	s.logger.Info(
		"Deleting expired sessions",
	)

	return s.repo.DeleteExpiredSessions(
		ctx,
	)
}

func (s *authSessionService) GetActiveSessionCount(
	ctx context.Context,
	userID uint,
) (int64, error) {

	return s.repo.GetActiveSessionCount(
		ctx,
		userID,
	)
}
