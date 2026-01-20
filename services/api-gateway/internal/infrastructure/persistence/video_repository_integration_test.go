package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/video-platform/services/api-gateway/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDatabase(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithStartupTimeout(60 * time.Second).
			WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get connection details
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Connect to database
	dsn := "host=" + host + " user=test_user password=test_password dbname=test_db port=" + port.Port() + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Create schema
	err = db.Exec("CREATE SCHEMA IF NOT EXISTS videos").Error
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&entities.Video{})
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		container.Terminate(ctx)
	}

	return db, cleanup
}

func TestVideoRepository_Create(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	video := &entities.Video{
		UserID:       1,
		Filename:     "test.mp4",
		OriginalPath: "uploads/test.mp4",
		Status:       entities.StatusPending,
		FPS:          30,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	err := repo.Create(ctx, video)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, video.ID)
	assert.NotZero(t, video.CreatedAt)
}

func TestVideoRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Create a video
	video := &entities.Video{
		UserID:       1,
		Filename:     "test.mp4",
		OriginalPath: "uploads/test.mp4",
		Status:       entities.StatusPending,
		FPS:          30,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	err := repo.Create(ctx, video)
	require.NoError(t, err)

	// Find by ID
	found, err := repo.FindByID(ctx, video.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, video.ID, found.ID)
	assert.Equal(t, video.UserID, found.UserID)
	assert.Equal(t, "test.mp4", found.Filename)
	assert.Equal(t, entities.StatusPending, found.Status)
}

func TestVideoRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Try to find non-existent video
	randomID := uuid.New()
	found, err := repo.FindByID(ctx, randomID)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "not found")
}

func TestVideoRepository_FindByUserID(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Create multiple videos for user 1
	for i := 0; i < 5; i++ {
		video := &entities.Video{
			UserID:       1,
			Filename:     "test" + string(rune(i)) + ".mp4",
			OriginalPath: "uploads/test.mp4",
			Status:       entities.StatusPending,
			FPS:          30,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		err := repo.Create(ctx, video)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Create video for user 2
	video := &entities.Video{
		UserID:       2,
		Filename:     "other.mp4",
		OriginalPath: "uploads/other.mp4",
		Status:       entities.StatusPending,
		FPS:          30,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	err := repo.Create(ctx, video)
	require.NoError(t, err)

	// Find videos for user 1
	videos, err := repo.FindByUserID(ctx, 1, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, videos, 5)
	// Should be ordered by created_at DESC
	for _, v := range videos {
		assert.Equal(t, int64(1), v.UserID)
	}
}

func TestVideoRepository_FindByUserID_Pagination(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Create 10 videos
	for i := 0; i < 10; i++ {
		video := &entities.Video{
			UserID:       1,
			Filename:     "test" + string(rune(i)) + ".mp4",
			OriginalPath: "uploads/test.mp4",
			Status:       entities.StatusPending,
			FPS:          30,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		err := repo.Create(ctx, video)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
	}

	// Get first page (5 items)
	page1, err := repo.FindByUserID(ctx, 1, 5, 0)
	assert.NoError(t, err)
	assert.Len(t, page1, 5)

	// Get second page (5 items)
	page2, err := repo.FindByUserID(ctx, 1, 5, 5)
	assert.NoError(t, err)
	assert.Len(t, page2, 5)

	// Ensure no overlap
	for _, v1 := range page1 {
		for _, v2 := range page2 {
			assert.NotEqual(t, v1.ID, v2.ID)
		}
	}
}

func TestVideoRepository_CountByUserID(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Create 3 videos for user 1
	for i := 0; i < 3; i++ {
		video := &entities.Video{
			UserID:       1,
			Filename:     "test.mp4",
			OriginalPath: "uploads/test.mp4",
			Status:       entities.StatusPending,
			FPS:          30,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		err := repo.Create(ctx, video)
		require.NoError(t, err)
	}

	// Create 2 videos for user 2
	for i := 0; i < 2; i++ {
		video := &entities.Video{
			UserID:       2,
			Filename:     "test.mp4",
			OriginalPath: "uploads/test.mp4",
			Status:       entities.StatusPending,
			FPS:          30,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		err := repo.Create(ctx, video)
		require.NoError(t, err)
	}

	// Count user 1 videos
	count, err := repo.CountByUserID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Count user 2 videos
	count, err = repo.CountByUserID(ctx, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Count non-existent user
	count, err = repo.CountByUserID(ctx, 999)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestVideoRepository_UpdateStatus(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Create a video
	video := &entities.Video{
		UserID:       1,
		Filename:     "test.mp4",
		OriginalPath: "uploads/test.mp4",
		Status:       entities.StatusPending,
		FPS:          30,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	err := repo.Create(ctx, video)
	require.NoError(t, err)

	// Update status to processing
	err = repo.UpdateStatus(ctx, video.ID, entities.StatusProcessing)
	assert.NoError(t, err)

	// Verify status was updated
	found, err := repo.FindByID(ctx, video.ID)
	assert.NoError(t, err)
	assert.Equal(t, entities.StatusProcessing, found.Status)

	// Update status to completed
	err = repo.UpdateStatus(ctx, video.ID, entities.StatusCompleted)
	assert.NoError(t, err)

	// Verify status was updated again
	found, err = repo.FindByID(ctx, video.ID)
	assert.NoError(t, err)
	assert.Equal(t, entities.StatusCompleted, found.Status)
}

func TestVideoRepository_UpdateStatus_NonExistent(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewVideoRepository(db)
	ctx := context.Background()

	// Try to update non-existent video
	randomID := uuid.New()
	err := repo.UpdateStatus(ctx, randomID, entities.StatusCompleted)

	// Should not error, but should not affect any rows
	assert.NoError(t, err)
}
