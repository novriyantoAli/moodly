package repository

import (
	"context"
	"testing"

	"github.com/novriyantoAli/moodly/internal/application/bill/dto"
	"github.com/novriyantoAli/moodly/internal/application/bill/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBillRepository_Create(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	defer testutil.CleanDB(db)

	logger := testutil.NewTestLogger(t)
	repo := NewBillRepository(db, logger)
	ctx := context.Background()

	t.Run("should create bill successfully", func(t *testing.T) {
		bill := testutil.CreateBillFixture()

		err := repo.Create(ctx, bill)
		assert.NoError(t, err)
		assert.NotEqual(t, bill.ID, [16]byte{})

		var dbBill entity.Bill
		err = db.First(&dbBill, "id = ?", bill.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, bill.SubscribeID, dbBill.SubscribeID)
		assert.Equal(t, bill.Amount, dbBill.Amount)
		assert.Equal(t, bill.BillMonth, dbBill.BillMonth)
		assert.Equal(t, bill.BillYear, dbBill.BillYear)
	})

	t.Run("should create multiple bills", func(t *testing.T) {
		// Clear any existing bills from previous subtest
		db.Exec("DELETE FROM bills")

		for i := 1; i <= 3; i++ {
			bill := testutil.CreateBillFixture()
			bill.SubscribeID = uint(i)
			err := repo.Create(ctx, bill)
			assert.NoError(t, err)
		}

		var count int64
		db.Model(&entity.Bill{}).Count(&count)
		assert.Equal(t, int64(3), count)
	})
}

func TestBillRepository_GetByID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	defer testutil.CleanDB(db)

	logger := testutil.NewTestLogger(t)
	repo := NewBillRepository(db, logger)
	ctx := context.Background()

	t.Run("should get bill by ID successfully", func(t *testing.T) {
		bill := testutil.CreateBillFixture()
		err := repo.Create(ctx, bill)
		require.NoError(t, err)

		foundBill, err := repo.GetByID(ctx, bill.ID)
		assert.NoError(t, err)
		assert.Equal(t, bill.ID, foundBill.ID)
		assert.Equal(t, bill.SubscribeID, foundBill.SubscribeID)
		assert.Equal(t, bill.Amount, foundBill.Amount)
	})

	t.Run("should return error when bill not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, testutil.CreateBillFixture().ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

func TestBillRepository_GetAll(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	defer testutil.CleanDB(db)

	logger := testutil.NewTestLogger(t)
	repo := NewBillRepository(db, logger)
	ctx := context.Background()

	t.Run("should get all bills with pagination", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			bill := testutil.CreateBillFixture()
			bill.SubscribeID = uint(i)
			err := repo.Create(ctx, bill)
			require.NoError(t, err)
		}

		filter := &dto.BillFilter{
			Page:     1,
			PageSize: 3,
		}

		bills, totalCount, err := repo.GetAll(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, bills, 3)
		assert.Equal(t, int64(5), totalCount)
	})

	t.Run("should filter bills by subscribe_id", func(t *testing.T) {
		db.Exec("DELETE FROM bills")

		bill1 := testutil.CreateBillFixture()
		bill1.SubscribeID = 10
		err := repo.Create(ctx, bill1)
		require.NoError(t, err)

		bill2 := testutil.CreateBillFixture()
		bill2.SubscribeID = 20
		err = repo.Create(ctx, bill2)
		require.NoError(t, err)

		filter := &dto.BillFilter{
			SubscribeID: 10,
		}

		bills, totalCount, err := repo.GetAll(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, bills, 1)
		assert.Equal(t, int64(1), totalCount)
		assert.Equal(t, uint(10), bills[0].SubscribeID)
	})

	t.Run("should filter bills by status", func(t *testing.T) {
		db.Exec("DELETE FROM bills")

		bill1 := testutil.CreateBillFixture()
		bill1.SubscribeID = 30
		bill1.Status = "unpaid"
		err := repo.Create(ctx, bill1)
		require.NoError(t, err)

		bill2 := testutil.CreateBillFixture()
		bill2.SubscribeID = 40
		bill2.Status = "paid"
		err = repo.Create(ctx, bill2)
		require.NoError(t, err)

		bill3 := testutil.CreateBillFixture()
		bill3.SubscribeID = 50
		bill3.Status = "unpaid"
		err = repo.Create(ctx, bill3)
		require.NoError(t, err)

		filter := &dto.BillFilter{
			Status: "unpaid",
		}

		bills, totalCount, err := repo.GetAll(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, bills, 2)
		assert.Equal(t, int64(2), totalCount)
		for _, bill := range bills {
			assert.Equal(t, entity.BillStatusUnpaid, bill.Status)
		}
	})

	t.Run("should handle pagination correctly", func(t *testing.T) {
		db.Exec("DELETE FROM bills")

		for i := 1; i <= 10; i++ {
			bill := testutil.CreateBillFixture()
			bill.SubscribeID = uint(100 + i)
			err := repo.Create(ctx, bill)
			require.NoError(t, err)
		}

		filter := &dto.BillFilter{
			Page:     1,
			PageSize: 5,
		}

		bills, totalCount, err := repo.GetAll(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, bills, 5)
		assert.Equal(t, int64(10), totalCount)

		filter.Page = 2
		bills, totalCount, err = repo.GetAll(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, bills, 5)
		assert.Equal(t, int64(10), totalCount)

		filter.Page = 3
		bills, totalCount, err = repo.GetAll(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, bills, 0)
	})
}

func TestBillRepository_ComplexScenario(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	defer testutil.CleanDB(db)

	logger := testutil.NewTestLogger(t)
	repo := NewBillRepository(db, logger)
	ctx := context.Background()

	t.Run("should handle complete workflow", func(t *testing.T) {
		// Create
		bill := testutil.CreateBillFixture()
		bill.SubscribeID = 111
		bill.Status = "unpaid"
		err := repo.Create(ctx, bill)
		require.NoError(t, err)

		// Read
		foundBill, err := repo.GetByID(ctx, bill.ID)
		require.NoError(t, err)
		assert.Equal(t, bill.SubscribeID, foundBill.SubscribeID)

		// Update via database
		foundBill.Status = "paid"
		err = db.Save(foundBill).Error
		assert.NoError(t, err)

		// Verify update
		updatedBill, err := repo.GetByID(ctx, bill.ID)
		require.NoError(t, err)
		assert.Equal(t, "paid", string(updatedBill.Status))

		// Get all with filter
		filter := &dto.BillFilter{
			SubscribeID: 111,
			Status:      "paid",
		}
		bills, count, err := repo.GetAll(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, bills, 1)
		assert.Equal(t, int64(1), count)

		// Delete via database
		err = db.Delete(&entity.Bill{}, "id = ?", bill.ID).Error
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, bill.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

func TestBillRepository_UpdateStatus(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	defer testutil.CleanDB(db)

	logger := testutil.NewTestLogger(t)
	repo := NewBillRepository(db, logger)
	ctx := context.Background()

	t.Run("should update bill status to paid", func(t *testing.T) {
		bill := testutil.CreateBillFixture()
		bill.Status = "unpaid"
		err := repo.Create(ctx, bill)
		require.NoError(t, err)

		// Update status to paid
		err = repo.UpdateStatus(ctx, bill.ID, "paid")
		assert.NoError(t, err)

		// Verify status was updated
		updatedBill, err := repo.GetByID(ctx, bill.ID)
		assert.NoError(t, err)
		assert.Equal(t, entity.BillStatusPaid, updatedBill.Status)
	})

	t.Run("should update bill status from paid to unpaid", func(t *testing.T) {
		db.Exec("DELETE FROM bills")

		bill := testutil.CreateBillFixture()
		bill.Status = "paid"
		err := repo.Create(ctx, bill)
		require.NoError(t, err)

		// Update status to unpaid
		err = repo.UpdateStatus(ctx, bill.ID, "unpaid")
		assert.NoError(t, err)

		// Verify status was updated
		updatedBill, err := repo.GetByID(ctx, bill.ID)
		assert.NoError(t, err)
		assert.Equal(t, "unpaid", string(updatedBill.Status))
	})

	t.Run("should handle update for non-existent bill", func(t *testing.T) {
		nonExistentID := testutil.CreateBillFixture().ID
		// No error is returned by GORM Update even if no rows are affected
		err := repo.UpdateStatus(ctx, nonExistentID, "paid")
		assert.NoError(t, err)
	})
}
