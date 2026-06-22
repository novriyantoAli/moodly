package usecase

import (
	"context"
	"errors"
	"testing"

	authDto "github.com/novriyantoAli/moodly/internal/application/auth/dto"
	authEntity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	securityDto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	userDto "github.com/novriyantoAli/moodly/internal/application/user/dto"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupLoginUseCaseWithMocks() (
	LoginUseCase,
	*testutil.MockUserService,
	*testutil.MockUserPasswordService,
	*testutil.MockAuthSessionService,
	*testutil.MockLoginAttemptService,
	*testutil.MockJWTService,
) {

	userSvc := &testutil.MockUserService{}
	passwordSvc := &testutil.MockUserPasswordService{}
	sessionSvc := &testutil.MockAuthSessionService{}
	attemptSvc := &testutil.MockLoginAttemptService{}

	logger := testutil.NewSilentLogger()

	jwtManager := &testutil.MockJWTService{}

	uc := NewLoginUseCase(
		userSvc,
		passwordSvc,
		sessionSvc,
		attemptSvc,
		jwtManager,
		logger,
	)

	return uc,
		userSvc,
		passwordSvc,
		sessionSvc,
		attemptSvc,
		jwtManager
}

func TestLoginUseCase_Execute(t *testing.T) {

	ctx := context.Background()

	t.Run(
		"should login successfully",
		func(t *testing.T) {

			uc,
				userSvc,
				passwordSvc,
				sessionSvc,
				attemptSvc,
				jwtManager := setupLoginUseCaseWithMocks()

			req := &authDto.LoginRequest{
				Username:  "admin@mail.com",
				Password:  "Password123",
				IPAddress: "127.0.0.1",
				UserAgent: "Chrome",
			}

			user := &userDto.UserResponse{
				ID:    1,
				Email: "admin@mail.com",
				Level: "ADMIN",
			}

			passwordSvc.
				On(
					"VerifyPassword",
					ctx,
					&securityDto.VerifyPasswordRequest{
						Username: req.Username,
						Password: req.Password,
					},
				).
				Return(true, nil)

			userSvc.
				On(
					"GetUserByEmail",
					ctx,
					req.Username,
				).
				Return(user, nil)

			jwtManager.
				On(
					"GenerateToken",
					user.ID,
					user.Email,
					user.Level,
				).
				Return("access-token", nil)

			jwtManager.
				On(
					"GenerateRefreshToken",
					user.ID,
					user.Email,
					user.Level,
				).
				Return("refresh-token", nil)

			sessionSvc.
				On(
					"CreateSession",
					ctx,
					mock.MatchedBy(func(session *authEntity.AuthSession) bool {

						return session.UserID == user.ID &&
							session.IPAddress == req.IPAddress &&
							session.UserAgent == req.UserAgent
					}),
				).
				Return(nil)

			attemptSvc.
				On(
					"CreateAttempt",
					ctx,
					mock.MatchedBy(func(attempt *authEntity.LoginAttempt) bool {

						return attempt.Success &&
							attempt.Username == user.Email
					}),
				).
				Return(nil)

			resp, err := uc.Execute(
				ctx,
				req,
			)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)

			userSvc.AssertExpectations(t)
			passwordSvc.AssertExpectations(t)
			sessionSvc.AssertExpectations(t)
			attemptSvc.AssertExpectations(t)

		},
	)

	t.Run(
		"should fail when password invalid",
		func(t *testing.T) {

			uc,
				_,
				passwordSvc,
				_,
				attemptSvc,
				_ := setupLoginUseCaseWithMocks()

			req := &authDto.LoginRequest{
				Username:  "admin@mail.com",
				Password:  "wrong-password",
				IPAddress: "127.0.0.1",
				UserAgent: "Chrome",
			}

			passwordSvc.
				On(
					"VerifyPassword",
					ctx,
					&securityDto.VerifyPasswordRequest{
						Username: req.Username,
						Password: req.Password,
					},
				).
				Return(false, nil)

			attemptSvc.
				On(
					"CreateAttempt",
					ctx,
					mock.MatchedBy(func(attempt *authEntity.LoginAttempt) bool {

						return !attempt.Success &&
							attempt.Username == req.Username
					}),
				).
				Return(nil)

			resp, err := uc.Execute(
				ctx,
				req,
			)

			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Equal(
				t,
				"invalid username or password",
				err.Error(),
			)

			passwordSvc.AssertExpectations(t)
			attemptSvc.AssertExpectations(t)
		},
	)

	t.Run(
		"should return error when verify password failed",
		func(t *testing.T) {

			uc,
				_,
				passwordSvc,
				_,
				_,
				_ := setupLoginUseCaseWithMocks()

			req := &authDto.LoginRequest{
				Username: "admin@mail.com",
				Password: "Password123",
			}

			passwordSvc.
				On(
					"VerifyPassword",
					ctx,
					mock.Anything,
				).
				Return(
					false,
					errors.New("database error"),
				)

			resp, err := uc.Execute(
				ctx,
				req,
			)

			assert.Error(t, err)
			assert.Nil(t, resp)

			passwordSvc.AssertExpectations(t)
		},
	)

	t.Run(
		"should return error when user not found",
		func(t *testing.T) {

			uc,
				userSvc,
				passwordSvc,
				_,
				_,
				_ := setupLoginUseCaseWithMocks()

			req := &authDto.LoginRequest{
				Username: "admin@mail.com",
				Password: "Password123",
			}

			passwordSvc.
				On(
					"VerifyPassword",
					ctx,
					mock.Anything,
				).
				Return(true, nil)

			userSvc.
				On(
					"GetUserByEmail",
					ctx,
					req.Username,
				).
				Return(
					nil,
					errors.New("user not found"),
				)

			resp, err := uc.Execute(
				ctx,
				req,
			)

			assert.Error(t, err)
			assert.Nil(t, resp)

			userSvc.AssertExpectations(t)
			passwordSvc.AssertExpectations(t)
		},
	)

	t.Run(
		"should return error when create session failed",
		func(t *testing.T) {

			uc,
				userSvc,
				passwordSvc,
				sessionSvc,
				_,
				jwtManager := setupLoginUseCaseWithMocks()

			req := &authDto.LoginRequest{
				Username: "admin@mail.com",
				Password: "Password123",
			}

			user := &userDto.UserResponse{
				ID:    1,
				Email: "admin@mail.com",
				Level: "ADMIN",
			}

			passwordSvc.
				On(
					"VerifyPassword",
					ctx,
					mock.Anything,
				).
				Return(true, nil)

			userSvc.
				On(
					"GetUserByEmail",
					ctx,
					req.Username,
				).
				Return(user, nil)

			jwtManager.
				On(
					"GenerateToken",
					user.ID,
					user.Email,
					user.Level,
				).
				Return("access-token", nil)

			jwtManager.
				On(
					"GenerateRefreshToken",
					user.ID,
					user.Email,
					user.Level,
				).
				Return("refresh-token", nil)

			sessionSvc.
				On(
					"CreateSession",
					ctx,
					mock.Anything,
				).
				Return(
					errors.New("session error"),
				)

			resp, err := uc.Execute(
				ctx,
				req,
			)

			assert.Error(t, err)
			assert.Nil(t, resp)

			userSvc.AssertExpectations(t)
			passwordSvc.AssertExpectations(t)
			sessionSvc.AssertExpectations(t)
		},
	)
}
