package repository

import (
	"context"
	"testing"

	testutil "github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/novriyantoAli/moodly/internal/application/security/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserOAuthRepository_Create(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserOAuthRepository(db, logger)
	ctx := context.Background()

	t.Run("should create oauth successfully", func(t *testing.T) {
		oauth := testutil.CreateUserOAuthFixture()
		oauth.UserID = 1
		oauth.Provider = "google"
		oauth.ProviderUserID = "google-123"

		err := repo.Create(ctx, oauth)

		assert.NoError(t, err)

		var dbOAuth entity.UserOAuth

		err = db.
			First(
				&dbOAuth,
				"user_id = ? AND provider = ?",
				oauth.UserID,
				oauth.Provider,
			).
			Error

		assert.NoError(t, err)
		assert.Equal(t, oauth.UserID, dbOAuth.UserID)
		assert.Equal(t, oauth.Provider, dbOAuth.Provider)
		assert.Equal(t, oauth.ProviderUserID, dbOAuth.ProviderUserID)
		assert.Equal(t, oauth.Email, dbOAuth.Email)
		assert.Equal(t, oauth.Name, dbOAuth.Name)
	})

	t.Run("should fail create duplicate provider user id", func(t *testing.T) {
		oauth1 := testutil.CreateUserOAuthFixture()
		oauth1.UserID = 2
		oauth1.Provider = "google"
		oauth1.ProviderUserID = "duplicate-id"

		oauth2 := testutil.CreateUserOAuthFixture()
		oauth2.UserID = 3
		oauth2.Provider = "google"
		oauth2.ProviderUserID = "duplicate-id"

		err1 := repo.Create(ctx, oauth1)
		err2 := repo.Create(ctx, oauth2)

		assert.NoError(t, err1)
		assert.Error(t, err2)
	})

	testutil.CleanDB(db)
}

func TestUserOAuthRepository_GetByProviderAndUserID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserOAuthRepository(db, logger)
	ctx := context.Background()

	t.Run("should get oauth by provider and provider user id", func(t *testing.T) {
		oauth := testutil.CreateUserOAuthFixture()
		oauth.UserID = 1
		oauth.Provider = "google"
		oauth.ProviderUserID = "google-123"

		err := repo.Create(ctx, oauth)
		require.NoError(t, err)

		found, err := repo.GetByProviderAndUserID(
			ctx,
			"google",
			"google-123",
		)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, oauth.UserID, found.UserID)
		assert.Equal(t, oauth.Provider, found.Provider)
		assert.Equal(t, oauth.ProviderUserID, found.ProviderUserID)
	})

	t.Run("should return error when oauth not found", func(t *testing.T) {
		found, err := repo.GetByProviderAndUserID(
			ctx,
			"google",
			"not-found",
		)

		assert.Error(t, err)
		assert.Nil(t, found)
	})

	testutil.CleanDB(db)
}

func TestUserOAuthRepository_GetByUserID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserOAuthRepository(db, logger)
	ctx := context.Background()

	t.Run("should get oauth list by user id", func(t *testing.T) {
		oauthGoogle := testutil.CreateUserOAuthFixture()
		oauthGoogle.UserID = 1
		oauthGoogle.Provider = "google"
		oauthGoogle.ProviderUserID = "google-123"

		oauthGithub := testutil.CreateUserOAuthFixture()
		oauthGithub.UserID = 1
		oauthGithub.Provider = "github"
		oauthGithub.ProviderUserID = "github-123"

		require.NoError(t, repo.Create(ctx, oauthGoogle))
		require.NoError(t, repo.Create(ctx, oauthGithub))

		result, err := repo.GetByUserID(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("should return empty list when user has no oauth", func(t *testing.T) {
		result, err := repo.GetByUserID(ctx, 999)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	testutil.CleanDB(db)
}

func TestUserOAuthRepository_Delete(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)
	repo := NewUserOAuthRepository(db, logger)
	ctx := context.Background()

	t.Run("should delete oauth by provider", func(t *testing.T) {
		oauth := testutil.CreateUserOAuthFixture()
		oauth.UserID = 1
		oauth.Provider = "google"
		oauth.ProviderUserID = "google-123"

		err := repo.Create(ctx, oauth)
		require.NoError(t, err)

		err = repo.Delete(
			ctx,
			oauth.UserID,
			oauth.Provider,
		)

		assert.NoError(t, err)

		found, err := repo.GetByProviderAndUserID(
			ctx,
			oauth.Provider,
			oauth.ProviderUserID,
		)

		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("should not fail when deleting non existing oauth", func(t *testing.T) {
		err := repo.Delete(
			ctx,
			999,
			"google",
		)

		assert.NoError(t, err)
	})

	testutil.CleanDB(db)
}
