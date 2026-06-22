package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/bill/dto"
	"github.com/novriyantoAli/moodly/internal/application/bill/entity"
	subscribeEntity "github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestBillService_CreateBill(t *testing.T) {
	ctx := context.Background()

	t.Run("should create bill successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		req := &dto.CreateBillRequest{
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
			Status:      "unpaid",
		}

		// Mock expectations
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Bill")).Return(nil).Run(func(args mock.Arguments) {
			bill := args.Get(1).(*entity.Bill)
			var zeroUUID uuid.UUID
			if bill.ID == zeroUUID {
				bill.ID = uuid.New()
			}
		})

		// When
		response, err := service.CreateBill(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.SubscribeID)
		assert.Equal(t, int64(150000), response.Amount)
		assert.Equal(t, uint(3), response.BillMonth)
		assert.Equal(t, "unpaid", response.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when subscribe_id is missing", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		req := &dto.CreateBillRequest{
			SubscribeID: 0, // Missing
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
		}

		// When
		response, err := service.CreateBill(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "subscribe_id is required", err.Error())
	})

	t.Run("should return error when amount is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		req := &dto.CreateBillRequest{
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      0, // Invalid
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
		}

		// When
		response, err := service.CreateBill(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "amount must be greater than 0", err.Error())
	})

	t.Run("should return error when bill_month is out of range", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		req := &dto.CreateBillRequest{
			SubscribeID: 1,
			BillMonth:   13, // Invalid
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
		}

		// When
		response, err := service.CreateBill(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "bill_month must be between 1 and 12", err.Error())
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		req := &dto.CreateBillRequest{
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
		}

		// Mock expectations
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Bill")).Return(errors.New("database error"))

		// When
		response, err := service.CreateBill(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestBillService_GetBillByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get bill by ID successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		billID := uuid.New()
		bill := &entity.Bill{
			ID:          billID,
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, billID).Return(bill, nil)

		// When
		response, err := service.GetBillByID(ctx, billID.String())

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, billID, response.ID)
		assert.Equal(t, uint(1), response.SubscribeID)
		assert.Equal(t, int64(150000), response.Amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when bill not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		billID := uuid.New()

		// Mock expectations
		mockRepo.On("GetByID", ctx, billID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetBillByID(ctx, billID.String())

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "bill not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when bill ID is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// When
		response, err := service.GetBillByID(ctx, "invalid-uuid")

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "invalid bill id format", err.Error())
		mockRepo.AssertNotCalled(t, "GetByID")
	})
}

func TestBillService_GetBills(t *testing.T) {
	ctx := context.Background()

	t.Run("should get bills with pagination", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		bills := []entity.Bill{
			{
				ID:          uuid.New(),
				SubscribeID: 1,
				BillMonth:   3,
				BillYear:    2026,
				Amount:      150000,
				DueDate:     time.Now().Add(30 * 24 * time.Hour),
				Status:      "unpaid",
				CreatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				SubscribeID: 2,
				BillMonth:   3,
				BillYear:    2026,
				Amount:      200000,
				DueDate:     time.Now().Add(30 * 24 * time.Hour),
				Status:      "paid",
				CreatedAt:   time.Now(),
			},
		}

		filter := &dto.BillFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(bills, int64(2), nil)

		// When
		response, err := service.GetBills(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, int64(2), response.TotalCount)
		assert.Equal(t, uint(1), response.Data[0].SubscribeID)
		assert.Equal(t, uint(2), response.Data[1].SubscribeID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should filter bills by subscribe_id", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		bills := []entity.Bill{
			{
				ID:          uuid.New(),
				SubscribeID: 1,
				BillMonth:   3,
				BillYear:    2026,
				Amount:      150000,
				DueDate:     time.Now().Add(30 * 24 * time.Hour),
				Status:      "unpaid",
				CreatedAt:   time.Now(),
			},
		}

		filter := &dto.BillFilter{
			SubscribeID: 1,
			Page:        1,
			PageSize:    10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(bills, int64(1), nil)

		// When
		response, err := service.GetBills(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, int64(1), response.TotalCount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should filter bills by status", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		bills := []entity.Bill{
			{
				ID:          uuid.New(),
				SubscribeID: 1,
				BillMonth:   3,
				BillYear:    2026,
				Amount:      150000,
				DueDate:     time.Now().Add(30 * 24 * time.Hour),
				Status:      "paid",
				CreatedAt:   time.Now(),
			},
		}

		filter := &dto.BillFilter{
			Status:   "paid",
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(bills, int64(1), nil)

		// When
		response, err := service.GetBills(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "paid", response.Data[0].Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should use default pagination when not provided", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		bills := []entity.Bill{}

		filter := &dto.BillFilter{}

		// Mock expectations - capture the filter to verify defaults are set
		mockRepo.On("GetAll", ctx, mock.MatchedBy(func(f *dto.BillFilter) bool {
			return f.Page == 1 && f.PageSize == 10
		})).Return(bills, int64(0), nil)

		// When
		response, err := service.GetBills(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.BillFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(nil, int64(0), errors.New("database error"))

		// When
		response, err := service.GetBills(ctx, filter)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestBillService_UpdateBillStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("should update bill status successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		billID := uuid.New()
		existingBill := &entity.Bill{
			ID:          billID,
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, billID).Return(existingBill, nil)
		mockRepo.On("UpdateStatus", ctx, billID, "paid").Return(nil)

		// When
		err := service.UpdateBillStatus(ctx, billID.String(), "paid")

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when bill not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		billID := uuid.New()

		// Mock expectations
		mockRepo.On("GetByID", ctx, billID).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.UpdateBillStatus(ctx, billID.String(), "paid")

		// Then
		assert.Error(t, err)
		assert.Equal(t, "bill not found", err.Error())
		mockRepo.AssertCalled(t, "GetByID", ctx, billID)
		mockRepo.AssertNotCalled(t, "UpdateStatus")
	})

	t.Run("should return error when status is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		billID := uuid.New()

		// When
		err := service.UpdateBillStatus(ctx, billID.String(), "invalid-status")

		// Then
		assert.Error(t, err)
		assert.Equal(t, "invalid status. allowed values: paid, unpaid, pending, overdue", err.Error())
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "UpdateStatus")
	})

	t.Run("should return error when bill ID is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// When
		err := service.UpdateBillStatus(ctx, "invalid-uuid", "paid")

		// Then
		assert.Error(t, err)
		assert.Equal(t, "invalid bill id format", err.Error())
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "UpdateStatus")
	})

	t.Run("should return error when repository update fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		billID := uuid.New()
		existingBill := &entity.Bill{
			ID:          billID,
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, billID).Return(existingBill, nil)
		mockRepo.On("UpdateStatus", ctx, billID, "paid").Return(errors.New("database error"))

		// When
		err := service.UpdateBillStatus(ctx, billID.String(), "paid")

		// Then
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should accept all valid status values", func(t *testing.T) {
		logger := testutil.NewSilentLogger()
		billID := uuid.New()
		existingBill := &entity.Bill{
			ID:          billID,
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}

		validStatuses := []string{"paid", "unpaid", "pending", "overdue"}

		for _, status := range validStatuses {
			// Reset mock for each iteration
			repo := &testutil.MockBillRepository{}
			mockSubRepo := &testutil.MockSubscribeRepository{}
			mockPublisher := &testutil.MockBillPublisher{}
			repo.On("GetByID", ctx, billID).Return(existingBill, nil)
			repo.On("UpdateStatus", ctx, billID, status).Return(nil)
			svc := NewBillService(repo, mockSubRepo, logger, mockPublisher)
			err := svc.UpdateBillStatus(ctx, billID.String(), status)

			// Then
			assert.NoError(t, err, "should accept status: %s", status)
		}
	})
}

func TestBillService_GenerateMonthlyBills(t *testing.T) {
	ctx := context.Background()

	t.Run("should generate monthly bills for all active subscribes", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		subscribes := []subscribeEntity.Subscriber{
			{
				ID:        1,
				Username:  "user1",
				CallName:  "User One",
				Plan:      "basic",
				Price:     150000,
				IsActive:  true,
				StartDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local),
			},
			{
				ID:        2,
				Username:  "user2",
				CallName:  "User Two",
				Plan:      "premium",
				Price:     300000,
				IsActive:  true,
				StartDate: time.Date(2025, 2, 10, 0, 0, 0, 0, time.Local),
			},
		}

		// Mock expectations
		mockSubRepo.On("GetActiveSubscribes", ctx).Return(subscribes, nil)
		mockRepo.On("Exists", ctx, uint(1), uint(3), uint(2026)).Return(false, nil)
		mockRepo.On("Exists", ctx, uint(2), uint(3), uint(2026)).Return(false, nil)
		mockPublisher.On("ScheduleBillPerSubscribe", mock.MatchedBy(func(req dto.CreateBillRequest) bool {
			return req.SubscribeID == 1 || req.SubscribeID == 2
		})).Return(nil)

		// When
		err := service.GenerateMonthlyBills(ctx, 3, 2026)

		// Then
		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
		mockPublisher.AssertNumberOfCalls(t, "ScheduleBillPerSubscribe", 2)
	})

	t.Run("should skip when bill already exists", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		subscribes := []subscribeEntity.Subscriber{
			{
				ID:        1,
				Username:  "user1",
				CallName:  "User One",
				Plan:      "basic",
				Price:     150000,
				IsActive:  true,
				StartDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local),
			},
		}

		// Mock expectations
		mockSubRepo.On("GetActiveSubscribes", ctx).Return(subscribes, nil)
		mockRepo.On("Exists", ctx, uint(1), uint(3), uint(2026)).Return(true, nil)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribe")

		// When
		err := service.GenerateMonthlyBills(ctx, 3, 2026)

		// Then
		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribe")
	})

	t.Run("should handle error when getting active subscribes", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockSubRepo.On("GetActiveSubscribes", ctx).Return(nil, errors.New("database error"))

		// When
		err := service.GenerateMonthlyBills(ctx, 3, 2026)

		// Then
		assert.Error(t, err)
		mockSubRepo.AssertExpectations(t)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribe")
	})

	t.Run("should continue when exists check fails for one subscribe", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		subscribes := []subscribeEntity.Subscriber{
			{
				ID:        1,
				Username:  "user1",
				Plan:      "basic",
				Price:     150000,
				IsActive:  true,
				StartDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local),
			},
			{
				ID:        2,
				Username:  "user2",
				Plan:      "premium",
				Price:     300000,
				IsActive:  true,
				StartDate: time.Date(2025, 2, 10, 0, 0, 0, 0, time.Local),
			},
		}

		// Mock expectations
		mockSubRepo.On("GetActiveSubscribes", ctx).Return(subscribes, nil)
		mockRepo.On("Exists", ctx, uint(1), uint(3), uint(2026)).Return(false, errors.New("exists check error"))
		mockRepo.On("Exists", ctx, uint(2), uint(3), uint(2026)).Return(false, nil)
		mockPublisher.On("ScheduleBillPerSubscribe", mock.MatchedBy(func(req dto.CreateBillRequest) bool {
			return req.SubscribeID == 2
		})).Return(nil)

		// When
		err := service.GenerateMonthlyBills(ctx, 3, 2026)

		// Then
		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
		// Should be called once for user2
		mockPublisher.AssertNumberOfCalls(t, "ScheduleBillPerSubscribe", 1)
	})

	t.Run("should continue when scheduling fails for one subscribe", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		subscribes := []subscribeEntity.Subscriber{
			{
				ID:        1,
				Username:  "user1",
				Plan:      "basic",
				Price:     150000,
				IsActive:  true,
				StartDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local),
			},
			{
				ID:        2,
				Username:  "user2",
				Plan:      "premium",
				Price:     300000,
				IsActive:  true,
				StartDate: time.Date(2025, 2, 10, 0, 0, 0, 0, time.Local),
			},
		}

		// Mock expectations
		mockSubRepo.On("GetActiveSubscribes", ctx).Return(subscribes, nil)
		mockRepo.On("Exists", ctx, uint(1), uint(3), uint(2026)).Return(false, nil)
		mockRepo.On("Exists", ctx, uint(2), uint(3), uint(2026)).Return(false, nil)
		mockPublisher.On("ScheduleBillPerSubscribe", mock.MatchedBy(func(req dto.CreateBillRequest) bool {
			return req.SubscribeID == 1
		})).Return(errors.New("scheduling error"))
		mockPublisher.On("ScheduleBillPerSubscribe", mock.MatchedBy(func(req dto.CreateBillRequest) bool {
			return req.SubscribeID == 2
		})).Return(nil)

		// When
		err := service.GenerateMonthlyBills(ctx, 3, 2026)

		// Then
		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
		// Should be called twice despite one error
		mockPublisher.AssertNumberOfCalls(t, "ScheduleBillPerSubscribe", 2)
	})

	t.Run("should handle no active subscribes", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockSubRepo.On("GetActiveSubscribes", ctx).Return([]subscribeEntity.Subscriber{}, nil)

		// When
		err := service.GenerateMonthlyBills(ctx, 3, 2026)

		// Then
		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribe")
	})
}

func TestBillService_QuickCountUnpaidBills(t *testing.T) {
	ctx := context.Background()

	t.Run("should count unpaid bills successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountUnpaidBills", ctx).Return(int64(5), nil)
		mockRepo.On("SumUnpaidBillsAmount", ctx).Return(int64(750000), nil)

		// When
		response, err := service.QuickCountUnpaidBills(ctx)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, int64(5), response.Count)
		assert.Equal(t, int64(750000), response.Amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return zero values when no unpaid bills", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountUnpaidBills", ctx).Return(int64(0), nil)
		mockRepo.On("SumUnpaidBillsAmount", ctx).Return(int64(0), nil)

		// When
		response, err := service.QuickCountUnpaidBills(ctx)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, int64(0), response.Count)
		assert.Equal(t, int64(0), response.Amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when count query fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountUnpaidBills", ctx).Return(int64(0), errors.New("database error"))

		// When
		response, err := service.QuickCountUnpaidBills(ctx)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertCalled(t, "CountUnpaidBills", ctx)
		mockRepo.AssertNotCalled(t, "SumUnpaidBillsAmount")
	})

	t.Run("should return error when sum query fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountUnpaidBills", ctx).Return(int64(5), nil)
		mockRepo.On("SumUnpaidBillsAmount", ctx).Return(int64(0), errors.New("database error"))

		// When
		response, err := service.QuickCountUnpaidBills(ctx)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestBillService_QuickCountOverdueBills(t *testing.T) {
	ctx := context.Background()

	t.Run("should count overdue bills successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountOverdueBills", ctx).Return(int64(3), nil)
		mockRepo.On("SumOverdueBillsAmount", ctx).Return(int64(450000), nil)

		// When
		response, err := service.QuickCountOverdueBills(ctx)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, int64(3), response.Count)
		assert.Equal(t, int64(450000), response.Amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return zero values when no overdue bills", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountOverdueBills", ctx).Return(int64(0), nil)
		mockRepo.On("SumOverdueBillsAmount", ctx).Return(int64(0), nil)

		// When
		response, err := service.QuickCountOverdueBills(ctx)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, int64(0), response.Count)
		assert.Equal(t, int64(0), response.Amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle high values correctly", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountOverdueBills", ctx).Return(int64(1000), nil)
		mockRepo.On("SumOverdueBillsAmount", ctx).Return(int64(150000000), nil)

		// When
		response, err := service.QuickCountOverdueBills(ctx)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, int64(1000), response.Count)
		assert.Equal(t, int64(150000000), response.Amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when count query fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountOverdueBills", ctx).Return(int64(0), errors.New("database error"))

		// When
		response, err := service.QuickCountOverdueBills(ctx)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertCalled(t, "CountOverdueBills", ctx)
		mockRepo.AssertNotCalled(t, "SumOverdueBillsAmount")
	})

	t.Run("should return error when sum query fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("CountOverdueBills", ctx).Return(int64(3), nil)
		mockRepo.On("SumOverdueBillsAmount", ctx).Return(int64(0), errors.New("database error"))
		// When
		response, err := service.QuickCountOverdueBills(ctx)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestBillService_ProcessUpdateBillStatusFromUnpaidToOverdue(t *testing.T) {
	ctx := context.Background()

	t.Run("should update unpaid bills to overdue when due date passed", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		pastDueDate := time.Now().Add(-10 * 24 * time.Hour)
		bills := []entity.Bill{
			{
				ID:          uuid.New(),
				SubscribeID: 1,
				BillMonth:   2,
				BillYear:    2026,
				Amount:      150000,
				DueDate:     pastDueDate,
				Status:      "unpaid",
				CreatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				SubscribeID: 2,
				BillMonth:   2,
				BillYear:    2026,
				Amount:      200000,
				DueDate:     pastDueDate,
				Status:      "unpaid",
				CreatedAt:   time.Now(),
			},
		}

		// Mock expectations
		mockRepo.On("GetActiveSubWithUnpaidBills", ctx).Return(bills, nil)
		mockPublisher.On("ScheduleBillPerSubscribeChangeFromUnpaidOverdue", mock.MatchedBy(func(bill entity.Bill) bool {
			return bill.SubscribeID == 1 || bill.SubscribeID == 2
		})).Return(nil)

		// When
		err := service.ProcessUpdateBillStatusFromUnpaidToOverdue(ctx, 2, 2026)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertNumberOfCalls(t, "ScheduleBillPerSubscribeChangeFromUnpaidOverdue", 2)
	})

	t.Run("should skip bills with future due dates", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		futureDueDate := time.Now().Add(10 * 24 * time.Hour)
		bills := []entity.Bill{
			{
				ID:          uuid.New(),
				SubscribeID: 1,
				BillMonth:   5,
				BillYear:    2026,
				Amount:      150000,
				DueDate:     futureDueDate,
				Status:      "unpaid",
				CreatedAt:   time.Now(),
			},
		}

		// Mock expectations
		mockRepo.On("GetActiveSubWithUnpaidBills", ctx).Return(bills, nil)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribeChangeFromUnpaidOverdue")

		// When
		err := service.ProcessUpdateBillStatusFromUnpaidToOverdue(ctx, 2, 2026)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribeChangeFromUnpaidOverdue")
	})

	t.Run("should continue when scheduling fails for one bill", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		pastDueDate := time.Now().Add(-10 * 24 * time.Hour)
		bills := []entity.Bill{
			{
				ID:          uuid.New(),
				SubscribeID: 1,
				BillMonth:   2,
				BillYear:    2026,
				Amount:      150000,
				DueDate:     pastDueDate,
				Status:      "unpaid",
				CreatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				SubscribeID: 2,
				BillMonth:   2,
				BillYear:    2026,
				Amount:      200000,
				DueDate:     pastDueDate,
				Status:      "unpaid",
				CreatedAt:   time.Now(),
			},
		}

		// Mock expectations
		mockRepo.On("GetActiveSubWithUnpaidBills", ctx).Return(bills, nil)
		mockPublisher.On("ScheduleBillPerSubscribeChangeFromUnpaidOverdue", bills[0]).Return(errors.New("scheduling error"))
		mockPublisher.On("ScheduleBillPerSubscribeChangeFromUnpaidOverdue", bills[1]).Return(nil)

		// When
		err := service.ProcessUpdateBillStatusFromUnpaidToOverdue(ctx, 2, 2026)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		// Should be called twice despite one error
		mockPublisher.AssertNumberOfCalls(t, "ScheduleBillPerSubscribeChangeFromUnpaidOverdue", 2)
	})

	t.Run("should handle empty bill list", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("GetActiveSubWithUnpaidBills", ctx).Return([]entity.Bill{}, nil)

		// When
		err := service.ProcessUpdateBillStatusFromUnpaidToOverdue(ctx, 2, 2026)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribeChangeFromUnpaidOverdue")
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		// Mock expectations
		mockRepo.On("GetActiveSubWithUnpaidBills", ctx).Return(nil, errors.New("database error"))

		// When
		err := service.ProcessUpdateBillStatusFromUnpaidToOverdue(ctx, 2, 2026)

		// Then
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockPublisher.AssertNotCalled(t, "ScheduleBillPerSubscribeChangeFromUnpaidOverdue")
	})

	t.Run("should handle mix of overdue and non-overdue bills", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		pastDueDate := time.Now().Add(-10 * 24 * time.Hour)
		futureDueDate := time.Now().Add(10 * 24 * time.Hour)
		overdueBill := entity.Bill{
			ID:          uuid.New(),
			SubscribeID: 1,
			BillMonth:   2,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     pastDueDate,
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}
		futureBill := entity.Bill{
			ID:          uuid.New(),
			SubscribeID: 2,
			BillMonth:   5,
			BillYear:    2026,
			Amount:      200000,
			DueDate:     futureDueDate,
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}
		bills := []entity.Bill{overdueBill, futureBill}

		// Mock expectations
		mockRepo.On("GetActiveSubWithUnpaidBills", ctx).Return(bills, nil)
		mockPublisher.On("ScheduleBillPerSubscribeChangeFromUnpaidOverdue", overdueBill).Return(nil)

		// When
		err := service.ProcessUpdateBillStatusFromUnpaidToOverdue(ctx, 2, 2026)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		// Should be called only once for the overdue bill
		mockPublisher.AssertNumberOfCalls(t, "ScheduleBillPerSubscribeChangeFromUnpaidOverdue", 1)
	})
}

func TestBillService_SumAmountByMonthYear(t *testing.T) {
	ctx := context.Background()

	t.Run("should sum amount by month year and status successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "paid",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "paid", uint(3), uint(2026)).Return(int64(500000), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(500000), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should sum all statuses when status is all", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "all",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "all", uint(3), uint(2026)).Return(int64(1500000), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(1500000), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return zero amount when no bills found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "paid",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "paid", uint(3), uint(2026)).Return(int64(0), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(0), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should use default month when month is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  0, // Invalid month
			Year:   2026,
			Status: "paid",
		}

		currentMonth := uint(time.Now().Month())

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "paid", currentMonth, uint(2026)).Return(int64(400000), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(400000), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should use default year when year is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   0, // Invalid year
			Status: "paid",
		}

		currentYear := uint(time.Now().Year())

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "paid", uint(3), currentYear).Return(int64(600000), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(600000), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should use default all status when status is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "invalid-status",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "all", uint(3), uint(2026)).Return(int64(700000), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(700000), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle high amount values", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "all",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "all", uint(3), uint(2026)).Return(int64(999999999999), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(999999999999), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "paid",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "paid", uint(3), uint(2026)).Return(int64(0), errors.New("database error"))

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.Error(t, err)
		assert.Equal(t, int64(0), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should filter by unpaid status", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "unpaid",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "unpaid", uint(3), uint(2026)).Return(int64(800000), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(800000), amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should filter by overdue status", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockBillRepository{}
		mockSubRepo := &testutil.MockSubscribeRepository{}
		mockPublisher := &testutil.MockBillPublisher{}
		logger := testutil.NewSilentLogger()
		service := NewBillService(mockRepo, mockSubRepo, logger, mockPublisher)

		filter := &dto.SumAmountBillFilter{
			Month:  3,
			Year:   2026,
			Status: "overdue",
		}

		// Mock expectations
		mockRepo.On("SumAmountByMonthYear", ctx, "overdue", uint(3), uint(2026)).Return(int64(200000), nil)

		// When
		amount, err := service.SumAmountByMonthYear(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(200000), amount)
		mockRepo.AssertExpectations(t)
	})
}
