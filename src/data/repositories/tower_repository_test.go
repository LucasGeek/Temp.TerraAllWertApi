package repositories

import (
	"context"
	"testing"

	"api/domain/entities"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Using SQLite for testing
	dsn := "file::memory:?cache=shared"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate for testing
	err = db.AutoMigrate(&entities.Tower{}, &entities.Floor{}, &entities.Apartment{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestTowerRepository_Create(t *testing.T) {
	// Note: This test would normally use a test database
	// For now, we'll test the structure
	t.Skip("Skipping integration test - requires test database setup")

	db := setupTestDB(t)
	repo := NewTowerRepository(db)
	ctx := context.Background()

	tower := &entities.Tower{
		Name:        "Test Tower",
		Description: stringPtr("Test Description"),
	}

	err := repo.Create(ctx, tower)
	assert.NoError(t, err)
	assert.NotEmpty(t, tower.ID)
}

func TestTowerRepository_GetByID(t *testing.T) {
	t.Skip("Skipping integration test - requires test database setup")

	db := setupTestDB(t)
	repo := NewTowerRepository(db)
	ctx := context.Background()

	// Create a tower first
	tower := &entities.Tower{
		Name:        "Test Tower",
		Description: stringPtr("Test Description"),
	}
	err := repo.Create(ctx, tower)
	assert.NoError(t, err)

	// Retrieve the tower
	retrievedTower, err := repo.GetByID(ctx, tower.ID)
	assert.NoError(t, err)
	assert.Equal(t, tower.Name, retrievedTower.Name)
	assert.Equal(t, tower.Description, retrievedTower.Description)
}

func TestTowerRepository_Update(t *testing.T) {
	t.Skip("Skipping integration test - requires test database setup")

	db := setupTestDB(t)
	repo := NewTowerRepository(db)
	ctx := context.Background()

	// Create a tower first
	tower := &entities.Tower{
		Name:        "Test Tower",
		Description: stringPtr("Test Description"),
	}
	err := repo.Create(ctx, tower)
	assert.NoError(t, err)

	// Update the tower
	tower.Name = "Updated Tower"
	tower.Description = stringPtr("Updated Description")
	err = repo.Update(ctx, tower)
	assert.NoError(t, err)

	// Retrieve and verify
	updatedTower, err := repo.GetByID(ctx, tower.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Tower", updatedTower.Name)
	assert.Equal(t, "Updated Description", *updatedTower.Description)
}

func TestTowerRepository_Delete(t *testing.T) {
	t.Skip("Skipping integration test - requires test database setup")

	db := setupTestDB(t)
	repo := NewTowerRepository(db)
	ctx := context.Background()

	// Create a tower first
	tower := &entities.Tower{
		Name:        "Test Tower",
		Description: stringPtr("Test Description"),
	}
	err := repo.Create(ctx, tower)
	assert.NoError(t, err)

	// Delete the tower
	err = repo.Delete(ctx, tower.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, tower.ID)
	assert.Error(t, err) // Should return an error since tower is deleted
}

func TestTowerRepository_ExistsByName(t *testing.T) {
	t.Skip("Skipping integration test - requires test database setup")

	db := setupTestDB(t)
	repo := NewTowerRepository(db)
	ctx := context.Background()

	// Test non-existing tower
	exists, err := repo.ExistsByName(ctx, "Non-existing Tower")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create a tower
	tower := &entities.Tower{
		Name:        "Test Tower",
		Description: stringPtr("Test Description"),
	}
	err = repo.Create(ctx, tower)
	assert.NoError(t, err)

	// Test existing tower
	exists, err = repo.ExistsByName(ctx, "Test Tower")
	assert.NoError(t, err)
	assert.True(t, exists)
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}