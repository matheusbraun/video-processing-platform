package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/video-platform/services/auth/internal/domain/entities"
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

	// Run migrations
	err = db.AutoMigrate(&entities.User{}, &entities.RefreshToken{})
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

func TestUserRepository_Create(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &entities.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	err := repo.Create(ctx, user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &entities.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find by email
	found, err := repo.FindByEmail(ctx, "test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "testuser", found.Username)
	assert.Equal(t, "test@example.com", found.Email)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to find non-existent user
	found, err := repo.FindByEmail(ctx, "nonexistent@example.com")

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &entities.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find by username
	found, err := repo.FindByUsername(ctx, "testuser")

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "testuser", found.Username)
}

func TestUserRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &entities.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find by ID
	found, err := repo.FindByID(ctx, user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "testuser", found.Username)
	assert.Equal(t, "test@example.com", found.Email)
}

func TestUserRepository_UniqueEmail(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create first user
	user1 := &entities.User{
		Username:     "user1",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create second user with same email
	user2 := &entities.User{
		Username:     "user2",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	err = repo.Create(ctx, user2)

	assert.Error(t, err)
}

func TestUserRepository_UniqueUsername(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create first user
	user1 := &entities.User{
		Username:     "testuser",
		Email:        "user1@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create second user with same username
	user2 := &entities.User{
		Username:     "testuser",
		Email:        "user2@example.com",
		PasswordHash: "hashed_password",
	}
	err = repo.Create(ctx, user2)

	assert.Error(t, err)
}
