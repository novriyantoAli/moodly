package service

import (
	"context"
	"errors"
	"testing"

	entity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
)

func setupLoginAttemptServiceWithMocks() (
	LoginAttemptService,
	*testutil.MockLoginAttemptRepository,
) {

	mockRepo := &testutil.MockLoginAttemptRepository{}
	logger := testutil.NewSilentLogger()

	service := NewLoginAttemptService(
		mockRepo,
		logger,
	)

	return service, mockRepo
}

func setupLoginAttemptService(
	mockRepo *testutil.MockLoginAttemptRepository,
) LoginAttemptService {

	logger := testutil.NewSilentLogger()

	return NewLoginAttemptService(
		mockRepo,
		logger,
	)
}

func TestLoginAttemptService_CreateAttempt(t *testing.T) {

	ctx := context.Background()

	t.Run(
		"should create login attempt successfully",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			userID := uint(1)

			attempt := &entity.LoginAttempt{
				UserID:   &userID,
				Username: "john",
				Success:  true,
			}

			mockRepo.
				On(
					"Create",
					ctx,
					attempt,
				).
				Return(nil)

			err := service.CreateAttempt(
				ctx,
				attempt,
			)

			assert.NoError(t, err)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should return error when repository fails",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			userID := uint(1)

			attempt := &entity.LoginAttempt{
				UserID:   &userID,
				Username: "john",
				Success:  true,
			}

			mockRepo.
				On(
					"Create",
					ctx,
					attempt,
				).
				Return(errors.New("database error"))

			err := service.CreateAttempt(
				ctx,
				attempt,
			)

			assert.Error(t, err)
			assert.Equal(
				t,
				"database error",
				err.Error(),
			)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestLoginAttemptService_GetAttemptsByUserID(
	t *testing.T,
) {

	ctx := context.Background()

	t.Run(
		"should get attempts by user id",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			userID := uint(1)

			attempts := []entity.LoginAttempt{
				{
					Username: "john",
					Success:  true,
				},
				{
					Username: "john",
					Success:  false,
				},
			}

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					userID,
				).
				Return(attempts, nil)

			result, err :=
				service.GetAttemptsByUserID(
					ctx,
					userID,
				)

			assert.NoError(t, err)
			assert.Len(t, result, 2)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should return error",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					uint(1),
				).
				Return(nil, errors.New("database error"))

			result, err :=
				service.GetAttemptsByUserID(
					ctx,
					1,
				)

			assert.Error(t, err)
			assert.Nil(t, result)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestLoginAttemptService_GetAttemptsByUsername(
	t *testing.T,
) {

	ctx := context.Background()

	t.Run(
		"should get attempts by username",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			attempts := []entity.LoginAttempt{
				{
					Username: "john",
					Success:  true,
				},
			}

			mockRepo.
				On(
					"GetByUsername",
					ctx,
					"john",
				).
				Return(attempts, nil)

			result, err :=
				service.GetAttemptsByUsername(
					ctx,
					"john",
				)

			assert.NoError(t, err)
			assert.Len(t, result, 1)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestLoginAttemptService_GetLatestAttemptByUserID(
	t *testing.T,
) {

	ctx := context.Background()

	t.Run(
		"should return latest attempt",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			userID := uint(1)

			attempts := []entity.LoginAttempt{
				{
					Username: "john",
					Success:  true,
				},
				{
					Username: "john",
					Success:  false,
				},
			}

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					userID,
				).
				Return(attempts, nil)

			result, err :=
				service.GetLatestAttemptByUserID(
					ctx,
					userID,
				)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Success)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should return nil when no attempts",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					uint(1),
				).
				Return([]entity.LoginAttempt{}, nil)

			result, err :=
				service.GetLatestAttemptByUserID(
					ctx,
					1,
				)

			assert.NoError(t, err)
			assert.Nil(t, result)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestLoginAttemptService_GetLatestAttemptByUsername(
	t *testing.T,
) {

	ctx := context.Background()

	t.Run(
		"should return latest attempt",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			attempts := []entity.LoginAttempt{
				{
					Username: "john",
					Success:  false,
				},
			}

			mockRepo.
				On(
					"GetByUsername",
					ctx,
					"john",
				).
				Return(attempts, nil)

			result, err :=
				service.GetLatestAttemptByUsername(
					ctx,
					"john",
				)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.False(t, result.Success)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestLoginAttemptService_GetFailedAttemptCountByUserID(
	t *testing.T,
) {

	ctx := context.Background()

	t.Run(
		"should count failed attempts",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			attempts := []entity.LoginAttempt{
				{Success: true},
				{Success: false},
				{Success: false},
			}

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					uint(1),
				).
				Return(attempts, nil)

			count, err :=
				service.GetFailedAttemptCountByUserID(
					ctx,
					1,
				)

			assert.NoError(t, err)
			assert.Equal(t, 2, count)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should return repository error",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					uint(1),
				).
				Return(nil, errors.New("database error"))

			count, err :=
				service.GetFailedAttemptCountByUserID(
					ctx,
					1,
				)

			assert.Error(t, err)
			assert.Equal(t, 0, count)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestLoginAttemptService_GetFailedAttemptCountByUsername(
	t *testing.T,
) {

	ctx := context.Background()

	t.Run(
		"should count failed attempts by username",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			attempts := []entity.LoginAttempt{
				{Success: false},
				{Success: false},
				{Success: true},
			}

			mockRepo.
				On(
					"GetByUsername",
					ctx,
					"john",
				).
				Return(attempts, nil)

			count, err :=
				service.GetFailedAttemptCountByUsername(
					ctx,
					"john",
				)

			assert.NoError(t, err)
			assert.Equal(t, 2, count)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should return repository error",
		func(t *testing.T) {

			service, mockRepo :=
				setupLoginAttemptServiceWithMocks()

			mockRepo.
				On(
					"GetByUsername",
					ctx,
					"john",
				).
				Return(nil, errors.New("database error"))

			count, err :=
				service.GetFailedAttemptCountByUsername(
					ctx,
					"john",
				)

			assert.Error(t, err)
			assert.Equal(t, 0, count)

			mockRepo.AssertExpectations(t)
		},
	)
}
