package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/jwt"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"
)

func newPaymentService(
	repo *testutil.MockPaymentRepository,
	billSvc *testutil.MockBillService,
	userSvc *testutil.MockUserService,
	generator *testutil.MockPaymentNumberGenerator,
	txManager *testutil.MockTransactionManager,
) PaymentService {

	return NewPaymentService(
		repo,
		billSvc,
		userSvc,
		generator,
		txManager,
		testutil.NewSilentLogger(),
	)
}

func TestPaymentService_GetClaims(t *testing.T) {

	t.Run("success", func(t *testing.T) {

		claims := &jwt.Claims{
			UserID: 1,
		}

		ctx := context.WithValue(
			context.Background(),
			jwt.ClaimsKey,
			claims,
		)

		result, err := GetClaims(ctx)

		assert.NoError(t, err)
		assert.Equal(t, claims, result)
	})

	t.Run("claims not found", func(t *testing.T) {

		_, err := GetClaims(
			context.Background(),
		)

		assert.Error(t, err)
		assert.EqualError(
			t,
			err,
			"claims not found in context",
		)
	})
}

func TestPaymentService_GetPaymentByID(
	t *testing.T,
) {

	t.Run("success", func(t *testing.T) {

		mockRepo := new(
			testutil.MockPaymentRepository,
		)

		svc := newPaymentService(
			mockRepo,
			nil,
			nil,
			nil,
			nil,
		)

		payment :=
			testutil.CreatePaymentFixture()

		payment.ID = 1

		mockRepo.
			On(
				"GetByID",
				mock.Anything,
				uint(1),
			).
			Return(
				payment,
				nil,
			)

		resp, err := svc.GetPaymentByID(
			context.Background(),
			1,
		)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(
			t,
			uint(1),
			resp.ID,
		)

		mockRepo.AssertExpectations(t)
	})

	t.Run("payment not found", func(t *testing.T) {

		mockRepo := new(
			testutil.MockPaymentRepository,
		)

		svc := newPaymentService(
			mockRepo,
			nil,
			nil,
			nil,
			nil,
		)

		mockRepo.
			On(
				"GetByID",
				mock.Anything,
				uint(1),
			).
			Return(
				nil,
				gorm.ErrRecordNotFound,
			)

		resp, err := svc.GetPaymentByID(
			context.Background(),
			1,
		)

		assert.Nil(t, resp)
		assert.EqualError(
			t,
			err,
			"payment not found",
		)
	})

	t.Run("repository error", func(t *testing.T) {

		mockRepo := new(
			testutil.MockPaymentRepository,
		)

		svc := newPaymentService(
			mockRepo,
			nil,
			nil,
			nil,
			nil,
		)

		mockRepo.
			On(
				"GetByID",
				mock.Anything,
				uint(1),
			).
			Return(
				nil,
				errors.New("db error"),
			)

		_, err := svc.GetPaymentByID(
			context.Background(),
			1,
		)

		assert.EqualError(
			t,
			err,
			"db error",
		)
	})
}

func TestPaymentService_GetPaymentByNumber(
	t *testing.T,
) {

	t.Run("success", func(t *testing.T) {

		mockRepo :=
			new(testutil.MockPaymentRepository)

		svc := newPaymentService(
			mockRepo,
			nil,
			nil,
			nil,
			nil,
		)

		payment :=
			testutil.CreatePaymentFixture()

		mockRepo.
			On(
				"GetByPaymentNumber",
				mock.Anything,
				"PAY-001",
			).
			Return(
				payment,
				nil,
			)

		resp, err :=
			svc.GetPaymentByNumber(
				context.Background(),
				"PAY-001",
			)

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {

		mockRepo :=
			new(testutil.MockPaymentRepository)

		svc := newPaymentService(
			mockRepo,
			nil,
			nil,
			nil,
			nil,
		)

		mockRepo.
			On(
				"GetByPaymentNumber",
				mock.Anything,
				"PAY-001",
			).
			Return(
				nil,
				gorm.ErrRecordNotFound,
			)

		_, err :=
			svc.GetPaymentByNumber(
				context.Background(),
				"PAY-001",
			)

		assert.EqualError(
			t,
			err,
			"payment not found",
		)
	})
}

func TestPaymentService_GetPayments(
	t *testing.T,
) {

	t.Run("default pagination", func(t *testing.T) {

		mockRepo :=
			new(testutil.MockPaymentRepository)

		svc := newPaymentService(
			mockRepo,
			nil,
			nil,
			nil,
			nil,
		)

		filter := &dto.PaymentFilter{}

		mockRepo.
			On(
				"GetAll",
				mock.Anything,
				filter,
			).
			Return(
				[]entity.Payment{},
				int64(0),
				nil,
			)

		resp, err :=
			svc.GetPayments(
				context.Background(),
				filter,
			)

		assert.NoError(t, err)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
	})
}
