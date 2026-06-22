package usecase

import (
	"context"
	"testing"

	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/entity"
	securityDto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	securityEntity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	userDto "github.com/novriyantoAli/moodly/internal/application/user/dto"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupGoogleLoginUseCase() (
	GoogleLoginUseCase,
	*testutil.MockUserService,
	*testutil.MockUserOAuthService,
	*testutil.MockOAuthService,
	*testutil.MockAuthSessionService,
	*testutil.MockLoginAttemptService,
	*testutil.MockJWTService,
) {

	logger := testutil.NewSilentLogger()

	mockUserSvc := &testutil.MockUserService{}
	mockUserOAuthSvc := &testutil.MockUserOAuthService{}
	mockOAuthSvc := &testutil.MockOAuthService{}
	mockSessionSvc := &testutil.MockAuthSessionService{}
	mockAttemptSvc := &testutil.MockLoginAttemptService{}
	mockTokenSvc := &testutil.MockJWTService{}

	uc := NewGoogleLoginUseCase(
		mockUserSvc,
		mockUserOAuthSvc,
		mockOAuthSvc,
		mockSessionSvc,
		mockAttemptSvc,
		mockTokenSvc,
		logger,
	)

	return uc,
		mockUserSvc,
		mockUserOAuthSvc,
		mockOAuthSvc,
		mockSessionSvc,
		mockAttemptSvc,
		mockTokenSvc
}

func TestGoogleLoginUseCase_Execute(t *testing.T) {

	t.Run("should login existing google user successfully", func(t *testing.T) {

		uc,
			mockUserSvc,
			mockUserOAuthSvc,
			mockOAuthSvc,
			mockSessionSvc,
			mockAttemptSvc,
			mockTokenSvc := setupGoogleLoginUseCase()

		req := &dto.GoogleLoginRequest{
			IDToken:   "valid-token",
			IPAddress: "127.0.0.1",
			UserAgent: "Mozilla",
		}

		mockOAuthSvc.On(
			"VerifyGoogleToken",
			mock.Anything,
			"valid-token",
		).Return(
			&securityDto.GoogleUserInfo{
				Subject: "google-123",
				Email:   "user@gmail.com",
				Name:    "John Doe",
			},
			nil,
		)

		mockUserOAuthSvc.On(
			"GetByProviderAndProviderUserID",
			mock.Anything,
			"google",
			"google-123",
		).Return(
			&securityEntity.UserOAuth{
				UserID: 1,
			},
			nil,
		)

		mockUserSvc.On(
			"GetUserByID",
			mock.Anything,
			uint(1),
		).Return(
			&userDto.UserResponse{
				ID:    1,
				Email: "user@gmail.com",
				Level: "member",
			},
			nil,
		)

		mockTokenSvc.On(
			"GenerateToken",
			uint(1),
			"user@gmail.com",
			"member",
		).Return(
			"access-token",
			nil,
		)

		mockTokenSvc.On(
			"GenerateRefreshToken",
			uint(1),
			"user@gmail.com",
			"member",
		).Return(
			"refresh-token",
			nil,
		)

		mockSessionSvc.On(
			"CreateSession",
			mock.Anything,
			mock.MatchedBy(func(s *entity.AuthSession) bool {
				return s.UserID == 1
			}),
		).Return(nil)

		mockAttemptSvc.On(
			"CreateAttempt",
			mock.Anything,
			mock.Anything,
		).Return(nil)

		resp, err := uc.Execute(
			context.Background(),
			req,
		)

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Equal(
			t,
			"access-token",
			resp.AccessToken,
		)

		assert.Equal(
			t,
			"refresh-token",
			resp.RefreshToken,
		)

		mockUserSvc.AssertExpectations(t)
		mockUserOAuthSvc.AssertExpectations(t)
		mockOAuthSvc.AssertExpectations(t)
		mockSessionSvc.AssertExpectations(t)
		mockAttemptSvc.AssertExpectations(t)
		mockTokenSvc.AssertExpectations(t)
	})
}
