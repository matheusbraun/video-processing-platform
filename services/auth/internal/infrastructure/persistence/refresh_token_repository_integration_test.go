package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/video-platform/services/auth/internal/domain/entities"
)

func TestRefreshTokenRepository_Create(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	token := &entities.RefreshToken{
		UserID:    1,
		Token:     "test_token_123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err := repo.Create(ctx, token)

	assert.NoError(t, err)
	assert.NotZero(t, token.ID)
	assert.NotZero(t, token.CreatedAt)
}

func TestRefreshTokenRepository_FindByToken(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	// Create a token
	token := &entities.RefreshToken{
		UserID:    1,
		Token:     "test_token_123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := repo.Create(ctx, token)
	require.NoError(t, err)

	// Find by token
	found, err := repo.FindByToken(ctx, "test_token_123")

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, int64(1), found.UserID)
	assert.Equal(t, "test_token_123", found.Token)
}

func TestRefreshTokenRepository_FindByToken_NotFound(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	// Try to find non-existent token
	found, err := repo.FindByToken(ctx, "nonexistent_token")

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "not found")
}

func TestRefreshTokenRepository_DeleteByToken(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	// Create a token
	token := &entities.RefreshToken{
		UserID:    1,
		Token:     "test_token_123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := repo.Create(ctx, token)
	require.NoError(t, err)

	// Delete by token
	err = repo.DeleteByToken(ctx, "test_token_123")
	assert.NoError(t, err)

	// Verify token was deleted
	found, err := repo.FindByToken(ctx, "test_token_123")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestRefreshTokenRepository_DeleteByToken_NonExistent(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	// Try to delete non-existent token
	err := repo.DeleteByToken(ctx, "nonexistent_token")

	// Should not error even if token doesn't exist
	assert.NoError(t, err)
}

func TestRefreshTokenRepository_DeleteByUserID(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	// Create multiple tokens for user 1
	for i := 0; i < 3; i++ {
		token := &entities.RefreshToken{
			UserID:    1,
			Token:     "user1_token_" + string(rune('a'+i)),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		err := repo.Create(ctx, token)
		require.NoError(t, err)
	}

	// Create token for user 2
	token := &entities.RefreshToken{
		UserID:    2,
		Token:     "user2_token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := repo.Create(ctx, token)
	require.NoError(t, err)

	// Delete all tokens for user 1
	err = repo.DeleteByUserID(ctx, 1)
	assert.NoError(t, err)

	// Verify user 1 tokens are deleted
	for i := 0; i < 3; i++ {
		found, err := repo.FindByToken(ctx, "user1_token_"+string(rune('a'+i)))
		assert.Error(t, err)
		assert.Nil(t, found)
	}

	// Verify user 2 token still exists
	found, err := repo.FindByToken(ctx, "user2_token")
	assert.NoError(t, err)
	assert.NotNil(t, found)
}

func TestRefreshTokenRepository_UniqueToken(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	// Create first token
	token1 := &entities.RefreshToken{
		UserID:    1,
		Token:     "duplicate_token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := repo.Create(ctx, token1)
	require.NoError(t, err)

	// Try to create second token with same token string
	token2 := &entities.RefreshToken{
		UserID:    2,
		Token:     "duplicate_token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = repo.Create(ctx, token2)

	assert.Error(t, err)
}

func TestRefreshTokenRepository_IsExpired(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	// Create expired token
	expiredToken := &entities.RefreshToken{
		UserID:    1,
		Token:     "expired_token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	err := repo.Create(ctx, expiredToken)
	require.NoError(t, err)

	// Create valid token
	validToken := &entities.RefreshToken{
		UserID:    1,
		Token:     "valid_token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = repo.Create(ctx, validToken)
	require.NoError(t, err)

	// Find and check expired token
	found, err := repo.FindByToken(ctx, "expired_token")
	assert.NoError(t, err)
	assert.True(t, found.IsExpired())

	// Find and check valid token
	found, err = repo.FindByToken(ctx, "valid_token")
	assert.NoError(t, err)
	assert.False(t, found.IsExpired())
}
