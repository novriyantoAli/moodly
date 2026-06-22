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

func TestAuthSessionRepository_Create(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should create session successfully", func(t *testing.T) {
		session := testutil.CreateAuthSessionFixture()
		session.UserID = 1

		err := repo.Create(ctx, session)

		assert.NoError(t, err)

		var dbSession entity.AuthSession

		err = db.
			First(&dbSession, session.ID).
			Error

		assert.NoError(t, err)
		assert.Equal(t, session.UserID, dbSession.UserID)
		assert.Equal(t, session.AccessToken, dbSession.AccessToken)
		assert.Equal(t, session.RefreshToken, dbSession.RefreshToken)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_GetByID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should get session by id", func(t *testing.T) {
		session := testutil.CreateAuthSessionFixture()

		err := repo.Create(ctx, session)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, session.ID)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, session.ID, found.ID)
	})

	t.Run("should return nil when session not found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, 999)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_GetByRefreshToken(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should get session by refresh token", func(t *testing.T) {
		session := testutil.CreateAuthSessionFixture()

		err := repo.Create(ctx, session)
		require.NoError(t, err)

		found, err := repo.GetByRefreshToken(
			ctx,
			session.RefreshToken,
		)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(
			t,
			session.RefreshToken,
			found.RefreshToken,
		)
	})

	t.Run("should return nil when token not found", func(t *testing.T) {
		found, err := repo.GetByRefreshToken(
			ctx,
			"unknown-token",
		)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_GetByUserID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should get sessions by user id", func(t *testing.T) {
		session1 := testutil.CreateAuthSessionFixture()
		session1.UserID = 1

		session2 := testutil.CreateAuthSessionFixture()
		session2.UserID = 1
		session2.RefreshToken = "refresh-token-2"

		require.NoError(t, repo.Create(ctx, session1))
		require.NoError(t, repo.Create(ctx, session2))

		sessions, err := repo.GetByUserID(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, sessions, 2)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_UpdateAccessToken(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should update access token", func(t *testing.T) {
		session := testutil.CreateAuthSessionFixture()

		require.NoError(t, repo.Create(ctx, session))

		newToken := "new-access-token"

		err = repo.UpdateAccessToken(
			ctx,
			session.ID,
			newToken,
		)

		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, session.ID)

		assert.NoError(t, err)
		assert.Equal(
			t,
			newToken,
			updated.AccessToken,
		)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_UpdateRefreshToken(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should update refresh token", func(t *testing.T) {
		session := testutil.CreateAuthSessionFixture()

		require.NoError(t, repo.Create(ctx, session))

		newRefresh := "new-refresh-token"
		newExpired := time.Now().Add(24 * time.Hour)

		err := repo.UpdateRefreshToken(
			ctx,
			session.ID,
			newRefresh,
			newExpired,
		)

		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, session.ID)

		assert.NoError(t, err)
		assert.Equal(t, newRefresh, updated.RefreshToken)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_Delete(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should delete session", func(t *testing.T) {
		session := testutil.CreateAuthSessionFixture()

		require.NoError(t, repo.Create(ctx, session))

		err := repo.Delete(ctx, session.ID)

		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, session.ID)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_DeleteByRefreshToken(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should delete by refresh token", func(t *testing.T) {
		session := testutil.CreateAuthSessionFixture()

		require.NoError(t, repo.Create(ctx, session))

		err := repo.DeleteByRefreshToken(
			ctx,
			session.RefreshToken,
		)

		assert.NoError(t, err)

		found, err := repo.GetByRefreshToken(
			ctx,
			session.RefreshToken,
		)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_DeleteByUserID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should delete all user sessions", func(t *testing.T) {
		session1 := testutil.CreateAuthSessionFixture()
		session1.UserID = 10

		session2 := testutil.CreateAuthSessionFixture()
		session2.UserID = 10
		session2.RefreshToken = "token-2"

		require.NoError(t, repo.Create(ctx, session1))
		require.NoError(t, repo.Create(ctx, session2))

		err := repo.DeleteByUserID(ctx, 10)

		assert.NoError(t, err)

		sessions, err := repo.GetByUserID(ctx, 10)

		assert.NoError(t, err)
		assert.Len(t, sessions, 0)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_DeleteExpiredSessions(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should delete expired sessions", func(t *testing.T) {
		expired := testutil.CreateAuthSessionFixture()
		expired.ExpiredAt = time.Now().UTC().Add(-1 * time.Hour)

		active := testutil.CreateAuthSessionFixture()
		active.RefreshToken = "active-token"
		active.ExpiredAt = time.Now().UTC().Add(24 * time.Hour)

		require.NoError(t, repo.Create(ctx, expired))
		require.NoError(t, repo.Create(ctx, active))

		err := repo.DeleteExpiredSessions(ctx)

		assert.NoError(t, err)

		expiredFound, _ := repo.GetByRefreshToken(
			ctx,
			expired.RefreshToken,
		)

		activeFound, _ := repo.GetByRefreshToken(
			ctx,
			active.RefreshToken,
		)

		assert.Nil(t, expiredFound)
		assert.NotNil(t, activeFound)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_GetActiveSessionCount(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should count active sessions only", func(t *testing.T) {
		active1 := testutil.CreateAuthSessionFixture()
		active1.UserID = 1
		active1.ExpiredAt = time.Now().UTC().Add(time.Hour)

		active2 := testutil.CreateAuthSessionFixture()
		active2.UserID = 1
		active2.RefreshToken = "active-2"
		active2.ExpiredAt = time.Now().UTC().Add(time.Hour)

		expired := testutil.CreateAuthSessionFixture()
		expired.UserID = 1
		expired.RefreshToken = "expired"
		expired.ExpiredAt = time.Now().UTC().Add(-time.Hour)

		require.NoError(t, repo.Create(ctx, active1))
		require.NoError(t, repo.Create(ctx, active2))
		require.NoError(t, repo.Create(ctx, expired))

		count, err := repo.GetActiveSessionCount(
			ctx,
			1,
		)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	testutil.CleanDB(db)
}

func TestAuthSessionRepository_GetOldestActiveSession(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewAuthSessionRepository(db, logger)

	ctx := context.Background()

	t.Run("should get oldest active session", func(t *testing.T) {
		oldest := testutil.CreateAuthSessionFixture()
		oldest.UserID = 1
		oldest.CreatedAt = time.Now().UTC().Add(-2 * time.Hour)

		newest := testutil.CreateAuthSessionFixture()
		newest.UserID = 1
		newest.RefreshToken = "newest-token"
		newest.CreatedAt = time.Now().UTC().Add(-1 * time.Hour)

		require.NoError(t, repo.Create(ctx, oldest))
		require.NoError(t, repo.Create(ctx, newest))

		found, err := repo.GetOldestActiveSession(
			ctx,
			1,
		)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(
			t,
			oldest.RefreshToken,
			found.RefreshToken,
		)
	})

	t.Run("should return nil when no active session", func(t *testing.T) {
		found, err := repo.GetOldestActiveSession(
			ctx,
			999,
		)

		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}
