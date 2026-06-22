package repository

import (
	"context"
	"testing"
	"time"

	testutil "github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/novriyantoAli/moodly/internal/application/security/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPasswordRepository_Create(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should create user password successfully", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1
		password.Username = "novri"

		err := repo.Create(ctx, password)

		assert.NoError(t, err)

		var dbPassword entity.UserPassword

		err = db.
			First(&dbPassword, "user_id = ?", password.UserID).
			Error

		assert.NoError(t, err)
		assert.Equal(t, password.UserID, dbPassword.UserID)
		assert.Equal(t, password.Username, dbPassword.Username)
		assert.Equal(t, password.PasswordHash, dbPassword.PasswordHash)
	})

	t.Run("should fail create duplicate user id", func(t *testing.T) {
		password1 := testutil.CreateUserPasswordFixture()
		password1.UserID = 2
		password1.Username = "user1"

		password2 := testutil.CreateUserPasswordFixture()
		password2.UserID = 2
		password2.Username = "user2"

		err1 := repo.Create(ctx, password1)
		err2 := repo.Create(ctx, password2)

		assert.NoError(t, err1)
		assert.Error(t, err2)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_GetByUserID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should get password by user id", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1
		password.Username = "novri"

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		found, err := repo.GetByUserID(ctx, password.UserID)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, password.UserID, found.UserID)
		assert.Equal(t, password.Username, found.Username)
	})

	t.Run("should return nil when not found", func(t *testing.T) {
		found, err := repo.GetByUserID(ctx, 999)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_GetByUsername(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should get password by username", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1
		password.Username = "admin"

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		found, err := repo.GetByUsername(ctx, "admin")

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "admin", found.Username)
	})

	t.Run("should return nil if username not found", func(t *testing.T) {
		found, err := repo.GetByUsername(ctx, "unknown")

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_UpdatePasswordHash(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should update password hash", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		newHash := "new_password_hash"

		err = repo.UpdatePasswordHash(
			ctx,
			password.UserID,
			newHash,
		)

		assert.NoError(t, err)

		updated, err := repo.GetByUserID(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)
		assert.Equal(t, newHash, updated.PasswordHash)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_IncrementFailedAttempt(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should increment failed attempt", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1
		password.FailedAttempt = 0

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		err = repo.IncrementFailedAttempt(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)

		updated, err := repo.GetByUserID(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)
		assert.Equal(t, 1, updated.FailedAttempt)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_ResetFailedAttempt(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should reset failed attempt", func(t *testing.T) {
		lockTime := time.Now().Add(15 * time.Minute)

		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1
		password.FailedAttempt = 3
		password.LockedUntil = &lockTime

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		err = repo.ResetFailedAttempt(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)

		updated, err := repo.GetByUserID(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)
		assert.Equal(t, 0, updated.FailedAttempt)
		assert.Nil(t, updated.LockedUntil)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_LockAccount(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should lock account", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		duration := 15 * time.Minute

		err = repo.LockAccount(
			ctx,
			password.UserID,
			duration,
		)

		assert.NoError(t, err)

		updated, err := repo.GetByUserID(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)
		assert.NotNil(t, updated.LockedUntil)
		assert.True(t, updated.LockedUntil.After(time.Now()))
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_Unlock(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should unlock account", func(t *testing.T) {
		lockTime := time.Now().Add(15 * time.Minute)

		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1
		password.LockedUntil = &lockTime

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		err = repo.Unlock(ctx, password.UserID)

		assert.NoError(t, err)

		updated, err := repo.GetByUserID(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)
		assert.Nil(t, updated.LockedUntil)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_UpdateLastLogin(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should update last login", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		loginTime := time.Now()

		err = repo.UpdateLastLogin(
			ctx,
			password.UserID,
			loginTime,
		)

		assert.NoError(t, err)

		updated, err := repo.GetByUserID(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)
		assert.NotNil(t, updated.LastLoginAt)
		assert.WithinDuration(
			t,
			loginTime,
			*updated.LastLoginAt,
			time.Second,
		)
	})

	testutil.CleanDB(db)
}

func TestUserPasswordRepository_Delete(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserPasswordRepository(db, logger)
	ctx := context.Background()

	t.Run("should delete password", func(t *testing.T) {
		password := testutil.CreateUserPasswordFixture()
		password.UserID = 1

		err := repo.Create(ctx, password)
		require.NoError(t, err)

		err = repo.Delete(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)

		found, err := repo.GetByUserID(
			ctx,
			password.UserID,
		)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}
