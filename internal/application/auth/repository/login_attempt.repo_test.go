package repository

import (
	"context"
	"testing"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/auth/entity"
	testutil "github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginAttemptRepository_Create(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewLoginAttemptRepository(db, logger)
	ctx := context.Background()

	t.Run("should create login attempt successfully", func(t *testing.T) {
		attempt := testutil.CreateLoginAttemptFixture()
		attempt.Username = "novri"

		err := repo.Create(ctx, attempt)

		assert.NoError(t, err)

		var dbAttempt entity.LoginAttempt

		err = db.
			First(&dbAttempt, "username = ?", attempt.Username).
			Error

		assert.NoError(t, err)
		assert.Equal(t, attempt.UserID, dbAttempt.UserID)
		assert.Equal(t, attempt.Username, dbAttempt.Username)
		assert.Equal(t, attempt.Success, dbAttempt.Success)
	})

	testutil.CleanDB(db)
}

func TestLoginAttemptRepository_GetByUserID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewLoginAttemptRepository(db, logger)
	ctx := context.Background()

	t.Run("should get login attempts by user id", func(t *testing.T) {
		attempt1 := testutil.CreateLoginAttemptFixture()
		attempt1.Username = "novri"

		attempt2 := testutil.CreateLoginAttemptFixture()
		attempt2.Username = "novri"

		err := repo.Create(ctx, attempt1)
		require.NoError(t, err)

		time.Sleep(time.Millisecond)

		err = repo.Create(ctx, attempt2)
		require.NoError(t, err)

		attempts, err := repo.GetByUserID(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, attempts, 2)

		// order by created_at desc
		assert.True(
			t,
			attempts[0].CreatedAt.After(attempts[1].CreatedAt) ||
				attempts[0].CreatedAt.Equal(attempts[1].CreatedAt),
		)
	})

	t.Run("should return empty slice when user not found", func(t *testing.T) {
		attempts, err := repo.GetByUserID(ctx, 999)

		assert.NoError(t, err)
		assert.Empty(t, attempts)
	})

	testutil.CleanDB(db)
}

func TestLoginAttemptRepository_GetByUsername(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewLoginAttemptRepository(db, logger)
	ctx := context.Background()

	t.Run("should get login attempts by username", func(t *testing.T) {
		attempt1 := testutil.CreateLoginAttemptFixture()
		attempt1.Username = "admin"

		attempt2 := testutil.CreateLoginAttemptFixture()
		uid := uint(2)
		attempt2.UserID = &uid
		attempt2.Username = "admin"

		err := repo.Create(ctx, attempt1)
		require.NoError(t, err)

		time.Sleep(time.Millisecond)

		err = repo.Create(ctx, attempt2)
		require.NoError(t, err)

		attempts, err := repo.GetByUsername(ctx, "admin")

		assert.NoError(t, err)
		assert.Len(t, attempts, 2)

		// order by created_at desc
		assert.True(
			t,
			attempts[0].CreatedAt.After(attempts[1].CreatedAt) ||
				attempts[0].CreatedAt.Equal(attempts[1].CreatedAt),
		)
	})

	t.Run("should return empty slice when username not found", func(t *testing.T) {
		attempts, err := repo.GetByUsername(ctx, "unknown")

		assert.NoError(t, err)
		assert.Empty(t, attempts)
	})

	testutil.CleanDB(db)
}
