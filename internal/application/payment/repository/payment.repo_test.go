package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupRepo(t *testing.T) (*gorm.DB, PaymentRepository) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)

	repo := NewPaymentRepository(
		db,
		logger,
	)

	return db, repo
}

func TestPaymentRepository_Create(t *testing.T) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	payment := testutil.CreatePaymentFixture()
	payment.ID = 0

	err := repo.Create(ctx, payment)

	require.NoError(t, err)
	assert.NotZero(t, payment.ID)

	var found entity.Payment

	err = db.First(&found, payment.ID).Error

	require.NoError(t, err)

	assert.Equal(t, payment.PaymentNumber, found.PaymentNumber)
	assert.Equal(t, payment.Amount, found.Amount)
	assert.Equal(t, payment.BillID, found.BillID)
}

func TestPaymentRepository_GetByID(t *testing.T) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	payment := testutil.CreatePaymentFixture()
	payment.ID = 0

	require.NoError(
		t,
		repo.Create(ctx, payment),
	)

	result, err := repo.GetByID(
		ctx,
		payment.ID,
	)

	require.NoError(t, err)

	assert.Equal(t, payment.ID, result.ID)
	assert.Equal(t, payment.PaymentNumber, result.PaymentNumber)
}

func TestPaymentRepository_GetByID_NotFound(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	_, err := repo.GetByID(
		context.Background(),
		99999,
	)

	assert.ErrorIs(
		t,
		err,
		gorm.ErrRecordNotFound,
	)
}

func TestPaymentRepository_GetByIDForUpdate(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	payment := testutil.CreatePaymentFixture()
	payment.ID = 0

	require.NoError(
		t,
		repo.Create(ctx, payment),
	)

	result, err := repo.GetByIDForUpdate(
		ctx,
		payment.ID,
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		payment.ID,
		result.ID,
	)
}

func TestPaymentRepository_GetByPaymentNumber(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	payment := testutil.CreatePaymentFixture()
	payment.ID = 0
	payment.PaymentNumber = "PAY-001"

	require.NoError(
		t,
		repo.Create(ctx, payment),
	)

	result, err := repo.GetByPaymentNumber(
		ctx,
		"PAY-001",
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		payment.ID,
		result.ID,
	)
}

func TestPaymentRepository_GetByBillID(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	billID := uuid.New()

	for i := 0; i < 3; i++ {

		payment := testutil.CreatePaymentFixture()

		payment.ID = 0
		payment.BillID = billID
		payment.PaymentNumber = fmt.Sprintf(
			"PAY-%d",
			i,
		)

		require.NoError(
			t,
			repo.Create(ctx, payment),
		)
	}

	results, err := repo.GetByBillID(
		ctx,
		billID.String(),
	)

	require.NoError(t, err)

	assert.Len(
		t,
		results,
		3,
	)
}

func TestPaymentRepository_GetAll(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	db.Exec("DELETE FROM payments")

	for i := 0; i < 5; i++ {

		payment := testutil.CreatePaymentFixture()

		payment.ID = 0
		payment.PaymentNumber = fmt.Sprintf(
			"PAY-%d",
			i,
		)

		require.NoError(
			t,
			repo.Create(ctx, payment),
		)
	}

	filter := &dto.PaymentFilter{
		Page:     1,
		PageSize: 3,
	}

	results, total, err := repo.GetAll(
		ctx,
		filter,
	)

	require.NoError(t, err)

	assert.Len(t, results, 3)
	assert.Equal(t, int64(5), total)
}

func TestPaymentRepository_Update(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	payment := testutil.CreatePaymentFixture()
	payment.ID = 0

	require.NoError(
		t,
		repo.Create(ctx, payment),
	)

	now := time.Now()

	payment.Status =
		entity.PaymentStatusCompleted

	payment.Description =
		"payment completed"

	payment.PaidAt = &now

	err := repo.Update(
		ctx,
		payment,
	)

	require.NoError(t, err)

	var found entity.Payment

	require.NoError(
		t,
		db.First(
			&found,
			payment.ID,
		).Error,
	)

	assert.Equal(
		t,
		entity.PaymentStatusCompleted,
		found.Status,
	)

	assert.Equal(
		t,
		"payment completed",
		found.Description,
	)

	assert.NotNil(
		t,
		found.PaidAt,
	)
}

func TestPaymentRepository_ExistsActivePaymentByBillID(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	billID := uuid.New()

	t.Run("should return true for pending payment", func(t *testing.T) {

		db.Exec("DELETE FROM payments")

		payment := testutil.CreatePaymentFixture()

		payment.ID = 0
		payment.BillID = billID
		payment.Status =
			entity.PaymentStatusPending

		require.NoError(
			t,
			repo.Create(ctx, payment),
		)

		exists, err :=
			repo.ExistsActivePaymentByBillID(
				ctx,
				billID.String(),
			)

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return true for completed payment", func(t *testing.T) {

		db.Exec("DELETE FROM payments")

		payment := testutil.CreatePaymentFixture()

		payment.ID = 0
		payment.BillID = billID
		payment.Status =
			entity.PaymentStatusCompleted

		require.NoError(
			t,
			repo.Create(ctx, payment),
		)

		exists, err :=
			repo.ExistsActivePaymentByBillID(
				ctx,
				billID.String(),
			)

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false for failed payment", func(t *testing.T) {

		db.Exec("DELETE FROM payments")

		payment := testutil.CreatePaymentFixture()

		payment.ID = 0
		payment.BillID = billID
		payment.Status =
			entity.PaymentStatusFailed

		require.NoError(
			t,
			repo.Create(ctx, payment),
		)

		exists, err :=
			repo.ExistsActivePaymentByBillID(
				ctx,
				billID.String(),
			)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should return false when payment not exists", func(t *testing.T) {

		db.Exec("DELETE FROM payments")

		exists, err :=
			repo.ExistsActivePaymentByBillID(
				ctx,
				uuid.New().String(),
			)

		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestPaymentRepository_GetByPaymentNumber_NotFound(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	_, err := repo.GetByPaymentNumber(
		context.Background(),
		"NOT-FOUND",
	)

	assert.ErrorIs(
		t,
		err,
		gorm.ErrRecordNotFound,
	)
}

func TestPaymentRepository_GetByIDForUpdate_NotFound(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	_, err := repo.GetByIDForUpdate(
		context.Background(),
		99999,
	)

	assert.ErrorIs(
		t,
		err,
		gorm.ErrRecordNotFound,
	)
}

func TestPaymentRepository_GetAll_FilterStatus(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	db.Exec("DELETE FROM payments")

	p1 := testutil.CreatePaymentFixture()
	p1.ID = 0
	p1.Status = entity.PaymentStatusPending
	p1.PaymentNumber = "PAY-PENDING"

	require.NoError(
		t,
		repo.Create(ctx, p1),
	)

	p2 := testutil.CreatePaymentFixture()
	p2.ID = 0
	p2.Status = entity.PaymentStatusCompleted
	p2.PaymentNumber = "PAY-COMPLETED"

	require.NoError(
		t,
		repo.Create(ctx, p2),
	)

	filter := &dto.PaymentFilter{
		Status: entity.PaymentStatusPending.String(),
	}

	results, total, err := repo.GetAll(
		ctx,
		filter,
	)

	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(
		t,
		entity.PaymentStatusPending,
		results[0].Status,
	)
}

func TestPaymentRepository_GetAll_FilterCurrency(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	db.Exec("DELETE FROM payments")

	p1 := testutil.CreatePaymentFixture()
	p1.ID = 0
	p1.Currency = "USD"
	p1.PaymentNumber = "PAY-USD"

	require.NoError(
		t,
		repo.Create(ctx, p1),
	)

	p2 := testutil.CreatePaymentFixture()
	p2.ID = 0
	p2.Currency = "EUR"
	p2.PaymentNumber = "PAY-EUR"

	require.NoError(
		t,
		repo.Create(ctx, p2),
	)

	filter := &dto.PaymentFilter{
		Currency: "USD",
	}

	results, total, err := repo.GetAll(
		ctx,
		filter,
	)

	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(
		t,
		"USD",
		results[0].Currency,
	)
}

func TestPaymentRepository_GetAll_FilterMethod(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	db.Exec("DELETE FROM payments")

	p1 := testutil.CreatePaymentFixture()
	p1.ID = 0
	p1.Method = entity.PaymentMethodCash
	p1.PaymentNumber = "PAY-CASH"

	require.NoError(
		t,
		repo.Create(ctx, p1),
	)

	p2 := testutil.CreatePaymentFixture()
	p2.ID = 0
	p2.Method = entity.PaymentMethodVirtualAccount
	p2.PaymentNumber = "PAY-VA"

	require.NoError(
		t,
		repo.Create(ctx, p2),
	)

	filter := &dto.PaymentFilter{
		Method: string(entity.PaymentMethodCash),
	}

	results, total, err := repo.GetAll(
		ctx,
		filter,
	)

	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(
		t,
		entity.PaymentMethodCash,
		results[0].Method,
	)
}

func TestPaymentRepository_GetAll_FilterBillID(
	t *testing.T,
) {

	db, repo := setupRepo(t)
	defer testutil.CleanDB(db)

	ctx := context.Background()

	db.Exec("DELETE FROM payments")

	billID1 := uuid.New()
	billID2 := uuid.New()

	p1 := testutil.CreatePaymentFixture()
	p1.ID = 0
	p1.BillID = billID1
	p1.PaymentNumber = "PAY-BILL-1"

	require.NoError(
		t,
		repo.Create(ctx, p1),
	)

	p2 := testutil.CreatePaymentFixture()
	p2.ID = 0
	p2.BillID = billID2
	p2.PaymentNumber = "PAY-BILL-2"

	require.NoError(
		t,
		repo.Create(ctx, p2),
	)

	filter := &dto.PaymentFilter{
		BillID: billID1.String(),
	}

	results, total, err := repo.GetAll(
		ctx,
		filter,
	)

	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(
		t,
		billID1,
		results[0].BillID,
	)
}
