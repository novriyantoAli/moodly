package repository

import (
	"context"
	"testing"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/security/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPINRepository_Create(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should create user PIN successfully", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1

		// When
		err := repo.Create(ctx, pin)

		// Then
		assert.NoError(t, err)

		// Verify pin was created in database
		var dbPin entity.UserPIN
		err = db.First(&dbPin, "user_id = ?", pin.UserID).Error
		assert.NoError(t, err)
		assert.Equal(t, pin.UserID, dbPin.UserID)
		assert.Equal(t, pin.PinHash, dbPin.PinHash)
		assert.Equal(t, pin.FailedAttempt, dbPin.FailedAttempt)
		assert.Equal(t, pin.LockedUntil, dbPin.LockedUntil)
	})

	t.Run("should fail to create security with duplicate user_id", func(t *testing.T) {
		// Given
		pin1 := testutil.CreateUserPINFixture()
		pin1.UserID = 2

		pin2 := testutil.CreateUserPINFixture()
		pin2.UserID = 2

		// When
		err1 := repo.Create(ctx, pin1)
		err2 := repo.Create(ctx, pin2)

		// Then
		assert.NoError(t, err1)
		assert.Error(t, err2) // Should fail due to primary key constraint
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserPINRepository_GetByUserID(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should get user pin by user_id successfully", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		// When
		foundPin, err := repo.GetByUserID(ctx, pin.UserID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, foundPin)
		assert.Equal(t, pin.UserID, foundPin.UserID)
		assert.Equal(t, pin.PinHash, foundPin.PinHash)
		assert.Equal(t, pin.FailedAttempt, foundPin.FailedAttempt)
	})

	t.Run("should return nil when user pin not found", func(t *testing.T) {
		// When
		foundPin, err := repo.GetByUserID(ctx, 999)

		// Then
		assert.NoError(t, err)
		assert.Nil(t, foundPin)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserPINRepository_UpdatePIN(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should update PIN successfully", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		newPinHash := "new_hashed_pin_value"

		// When
		err = repo.UpdatePIN(ctx, pin.UserID, newPinHash)

		// Then
		assert.NoError(t, err)

		// Verify PIN was updated
		updatedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.Equal(t, newPinHash, updatedPin.PinHash)
	})

	t.Run("should not fail when updating PIN for non-existent user", func(t *testing.T) {
		// When
		err := repo.UpdatePIN(ctx, 999, "new_pin_hash")

		// Then
		assert.NoError(t, err) // GORM doesn't error on update with no rows affected
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserPINRepository_IncrementFailedAttempt(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should increment failed attempt successfully", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1
		pin.FailedAttempt = 0
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		// When
		err = repo.IncrementFailedAttempt(ctx, pin.UserID)

		// Then
		assert.NoError(t, err)

		// Verify failed attempt was incremented
		updatedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.Equal(t, 1, updatedPin.FailedAttempt)
	})

	t.Run("should increment failed attempt multiple times", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 2
		pin.FailedAttempt = 0
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		// When
		for i := 0; i < 3; i++ {
			err = repo.IncrementFailedAttempt(ctx, pin.UserID)
			require.NoError(t, err)
		}

		// Then
		updatedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.Equal(t, 3, updatedPin.FailedAttempt)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserPINRepository_ResetFailedAttempt(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should reset failed attempt to zero", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1
		pin.FailedAttempt = 3
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		// When
		err = repo.ResetFailedAttempt(ctx, pin.UserID)

		// Then
		assert.NoError(t, err)

		// Verify failed attempt was reset and locked_until cleared
		updatedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.Equal(t, 0, updatedPin.FailedAttempt)
		assert.Nil(t, updatedPin.LockedUntil)
	})

	t.Run("should clear locked_until when resetting failed attempt", func(t *testing.T) {
		// Given
		lockedUntil := time.Now().Add(15 * time.Minute)
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 2
		pin.FailedAttempt = 3
		pin.LockedUntil = &lockedUntil
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		// When
		err = repo.ResetFailedAttempt(ctx, pin.UserID)

		// Then
		assert.NoError(t, err)

		// Verify locked_until was cleared
		updatedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.Nil(t, updatedPin.LockedUntil)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserPINRepository_LockAccount(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should lock account successfully", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		duration := 15 * time.Minute

		// When
		err = repo.LockAccount(ctx, pin.UserID, duration)

		// Then
		assert.NoError(t, err)

		// Verify account is locked
		updatedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedPin.LockedUntil)
		assert.True(t, updatedPin.LockedUntil.After(time.Now()))
		assert.True(t, updatedPin.LockedUntil.Before(time.Now().Add(duration+1*time.Second)))
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserPINRepository_Unlock(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should unlock account successfully", func(t *testing.T) {
		// Given
		lockedUntil := time.Now().Add(15 * time.Minute)
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1
		pin.LockedUntil = &lockedUntil
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		// When
		err = repo.Unlock(ctx, pin.UserID)

		// Then
		assert.NoError(t, err)

		// Verify account is unlocked
		updatedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.Nil(t, updatedPin.LockedUntil)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserPINRepository_Delete(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserPINRepository(db, logger)
	ctx := context.Background()

	t.Run("should delete user pin successfully", func(t *testing.T) {
		// Given
		pin := testutil.CreateUserPINFixture()
		pin.UserID = 1
		err := repo.Create(ctx, pin)
		require.NoError(t, err)

		// When
		err = repo.Delete(ctx, pin.UserID)

		// Then
		assert.NoError(t, err)

		// Verify pin was deleted
		deletedPin, err := repo.GetByUserID(ctx, pin.UserID)
		assert.NoError(t, err)
		assert.Nil(t, deletedPin)
	})

	t.Run("should not fail when deleting non-existent security record", func(t *testing.T) {
		// When
		err := repo.Delete(ctx, 999)

		// Then
		assert.NoError(t, err) // GORM doesn't error on delete with no rows affected
	})

	// Cleanup
	testutil.CleanDB(db)
}
