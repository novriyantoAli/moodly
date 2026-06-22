package service

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
)

func setupAuthSessionServiceWithMocks() (
	AuthSessionService,
	*testutil.MockAuthSessionRepository,
) {

	mockRepo := &testutil.MockAuthSessionRepository{}
	logger := testutil.NewSilentLogger()

	service := NewAuthSessionService(
		mockRepo,
		logger,
	)

	return service, mockRepo
}

func setupAuthSessionService(
	mockRepo *testutil.MockAuthSessionRepository,
) AuthSessionService {

	logger := testutil.NewSilentLogger()

	return NewAuthSessionService(
		mockRepo,
		logger,
	)
}

func TestAuthSessionService_CreateSession(t *testing.T) {

	ctx := context.Background()

	t.Run(
		"should create session successfully when active session below limit",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			session := &entity.AuthSession{
				UserID:       1,
				AccessToken:  "access",
				RefreshToken: "refresh",
			}

			mockRepo.
				On(
					"GetActiveSessionCount",
					ctx,
					uint(1),
				).
				Return(int64(1), nil)

			mockRepo.
				On(
					"Create",
					ctx,
					session,
				).
				Return(nil)

			err := service.CreateSession(
				ctx,
				session,
			)

			assert.NoError(t, err)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should delete oldest session when limit reached",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			session := &entity.AuthSession{
				UserID:       1,
				AccessToken:  "access",
				RefreshToken: "refresh",
			}

			oldest := &entity.AuthSession{
				ID:     10,
				UserID: 1,
			}

			mockRepo.
				On(
					"GetActiveSessionCount",
					ctx,
					uint(1),
				).
				Return(int64(2), nil)

			mockRepo.
				On(
					"GetOldestActiveSession",
					ctx,
					uint(1),
				).
				Return(oldest, nil)

			mockRepo.
				On(
					"Delete",
					ctx,
					uint(10),
				).
				Return(nil)

			mockRepo.
				On(
					"Create",
					ctx,
					session,
				).
				Return(nil)

			err := service.CreateSession(
				ctx,
				session,
			)

			assert.NoError(t, err)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should return error when get active session count fails",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			session := &entity.AuthSession{
				UserID: 1,
			}

			mockRepo.
				On(
					"GetActiveSessionCount",
					ctx,
					uint(1),
				).
				Return(int64(0), errors.New("database error"))

			err := service.CreateSession(
				ctx,
				session,
			)

			assert.Error(t, err)
			assert.Equal(
				t,
				"database error",
				err.Error(),
			)
		},
	)
}

func TestAuthSessionService_GetSessionByID(t *testing.T) {

	ctx := context.Background()

	service, mockRepo :=
		setupAuthSessionServiceWithMocks()

	session := &entity.AuthSession{
		ID: 1,
	}

	mockRepo.
		On(
			"GetByID",
			ctx,
			uint(1),
		).
		Return(session, nil)

	result, err := service.GetSessionByID(
		ctx,
		1,
	)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)

	mockRepo.AssertExpectations(t)
}

func TestAuthSessionService_GetSessionByRefreshToken(t *testing.T) {

	ctx := context.Background()

	t.Run(
		"should return session when token valid",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			session := &entity.AuthSession{
				ID:           1,
				RefreshToken: "refresh",
				ExpiredAt:    time.Now().Add(time.Hour),
			}

			mockRepo.
				On(
					"GetByRefreshToken",
					ctx,
					"refresh",
				).
				Return(session, nil)

			result, err :=
				service.GetSessionByRefreshToken(
					ctx,
					"refresh",
				)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should return nil when token expired",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			session := &entity.AuthSession{
				ID:           1,
				RefreshToken: "refresh",
				ExpiredAt:    time.Now().Add(-time.Hour),
			}

			mockRepo.
				On(
					"GetByRefreshToken",
					ctx,
					"refresh",
				).
				Return(session, nil)

			result, err :=
				service.GetSessionByRefreshToken(
					ctx,
					"refresh",
				)

			assert.NoError(t, err)
			assert.Nil(t, result)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestAuthSessionService_GetUserSessions(t *testing.T) {

	ctx := context.Background()

	service, mockRepo :=
		setupAuthSessionServiceWithMocks()

	sessions := []entity.AuthSession{
		{
			ID:     1,
			UserID: 1,
		},
		{
			ID:     2,
			UserID: 1,
		},
	}

	mockRepo.
		On(
			"GetByUserID",
			ctx,
			uint(1),
		).
		Return(sessions, nil)

	result, err := service.GetUserSessions(
		ctx,
		1,
	)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	mockRepo.AssertExpectations(t)
}

func TestAuthSessionService_RefreshSession(t *testing.T) {

	ctx := context.Background()

	t.Run(
		"should refresh token successfully",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			session := &entity.AuthSession{
				ID:           1,
				RefreshToken: "old-refresh",
				ExpiredAt:    time.Now().Add(time.Hour),
			}

			expiredAt := time.Now().Add(
				24 * time.Hour,
			)

			mockRepo.
				On(
					"GetByRefreshToken",
					ctx,
					"old-refresh",
				).
				Return(session, nil)

			mockRepo.
				On(
					"UpdateAccessToken",
					ctx,
					uint(1),
					"new-access",
				).
				Return(nil)

			mockRepo.
				On(
					"UpdateRefreshToken",
					ctx,
					uint(1),
					"new-refresh",
					expiredAt,
				).
				Return(nil)

			err := service.RefreshSession(
				ctx,
				"old-refresh",
				"new-access",
				"new-refresh",
				expiredAt,
			)

			assert.NoError(t, err)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should fail when session not found",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			mockRepo.
				On(
					"GetByRefreshToken",
					ctx,
					"invalid",
				).
				Return(nil, nil)

			err := service.RefreshSession(
				ctx,
				"invalid",
				"access",
				"refresh",
				time.Now(),
			)

			assert.Error(t, err)
			assert.Equal(
				t,
				"session not found",
				err.Error(),
			)
		},
	)

	t.Run(
		"should fail when refresh token expired",
		func(t *testing.T) {

			service, mockRepo :=
				setupAuthSessionServiceWithMocks()

			session := &entity.AuthSession{
				ID:        1,
				ExpiredAt: time.Now().Add(-time.Hour),
			}

			mockRepo.
				On(
					"GetByRefreshToken",
					ctx,
					"expired",
				).
				Return(session, nil)

			err := service.RefreshSession(
				ctx,
				"expired",
				"access",
				"refresh",
				time.Now(),
			)

			assert.Error(t, err)
			assert.Equal(
				t,
				"refresh token expired",
				err.Error(),
			)
		},
	)
}

func TestAuthSessionService_Logout(t *testing.T) {

	ctx := context.Background()

	service, mockRepo :=
		setupAuthSessionServiceWithMocks()

	mockRepo.
		On(
			"Delete",
			ctx,
			uint(1),
		).
		Return(nil)

	err := service.Logout(
		ctx,
		1,
	)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthSessionService_LogoutByRefreshToken(t *testing.T) {

	ctx := context.Background()

	service, mockRepo :=
		setupAuthSessionServiceWithMocks()

	mockRepo.
		On(
			"DeleteByRefreshToken",
			ctx,
			"refresh-token",
		).
		Return(nil)

	err := service.LogoutByRefreshToken(
		ctx,
		"refresh-token",
	)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthSessionService_LogoutAllUserSessions(t *testing.T) {

	ctx := context.Background()

	service, mockRepo :=
		setupAuthSessionServiceWithMocks()

	mockRepo.
		On(
			"DeleteByUserID",
			ctx,
			uint(1),
		).
		Return(nil)

	err := service.LogoutAllUserSessions(
		ctx,
		1,
	)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthSessionService_DeleteExpiredSessions(t *testing.T) {

	ctx := context.Background()

	service, mockRepo :=
		setupAuthSessionServiceWithMocks()

	mockRepo.
		On(
			"DeleteExpiredSessions",
			ctx,
		).
		Return(nil)

	err := service.DeleteExpiredSessions(
		ctx,
	)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthSessionService_GetActiveSessionCount(t *testing.T) {

	ctx := context.Background()

	service, mockRepo :=
		setupAuthSessionServiceWithMocks()

	mockRepo.
		On(
			"GetActiveSessionCount",
			ctx,
			uint(1),
		).
		Return(int64(2), nil)

	count, err := service.GetActiveSessionCount(
		ctx,
		1,
	)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	mockRepo.AssertExpectations(t)
}
