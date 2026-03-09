package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseService handles local SQLite cache operations
type DatabaseService struct {
	db *gorm.DB
}

// NewDatabaseService creates a new database service
func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database with GORM
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	service := &DatabaseService{db: db}

	// Auto-migrate schema
	if err := service.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("✅ Local cache database initialized successfully")

	return service, nil
}

// Close closes the database connection
func (ds *DatabaseService) Close() error {
	sqlDB, err := ds.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// initSchema creates all necessary tables for local cache using GORM AutoMigrate
func (ds *DatabaseService) initSchema() error {
	// GORM will automatically create tables based on the model
	return ds.db.AutoMigrate(&models.CachedRepository{})
}

// CacheRepository stores a repository in the local cache
func (ds *DatabaseService) CacheRepository(repo *models.Repository) error {
	now := time.Now()

	cachedRepo := &models.CachedRepository{
		RepositoryID:    repo.ID,
		Name:            repo.Name,
		GitHubURL:       repo.GitHubURL,
		Owner:           repo.Owner,
		LastCommitTime:  repo.LastCommitTime,
		LastCommitBy:    repo.LastCommitBy,
		DefaultBranch:   repo.DefaultBranch,
		EnvironmentName: repo.EnvironmentName,
		JiraProjectKey:  repo.JiraProjectKey,
		CachedAt:        now,
		UpdatedAt:       now,
	}

	// Use GORM's Save method (inserts or updates based on primary key)
	result := ds.db.Where("repository_id = ?", repo.ID).
		Assign(cachedRepo).
		FirstOrCreate(cachedRepo)

	if result.Error != nil {
		return fmt.Errorf("failed to cache repository: %w", result.Error)
	}

	return nil
}

// GetCachedRepository retrieves a repository from cache by name
func (ds *DatabaseService) GetCachedRepository(name string) (*models.CachedRepository, error) {
	var repo models.CachedRepository

	result := ds.db.Where("name = ?", name).First(&repo)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Not found in cache
		}
		return nil, fmt.Errorf("failed to get cached repository: %w", result.Error)
	}

	return &repo, nil
}

// GetCachedRepositoryByID retrieves a repository from cache by repository ID
func (ds *DatabaseService) GetCachedRepositoryByID(repoID int64) (*models.CachedRepository, error) {
	var repo models.CachedRepository

	result := ds.db.Where("repository_id = ?", repoID).First(&repo)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Not found in cache
		}
		return nil, fmt.Errorf("failed to get cached repository by ID: %w", result.Error)
	}

	return &repo, nil
}

// GetAllCachedRepositories retrieves all repositories from cache
func (ds *DatabaseService) GetAllCachedRepositories() ([]*models.CachedRepository, error) {
	var repos []*models.CachedRepository

	result := ds.db.Order("name").Find(&repos)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query cached repositories: %w", result.Error)
	}

	return repos, nil
}

// IsCacheStale checks if cache is older than the specified duration
func (ds *DatabaseService) IsCacheStale(name string, maxAge time.Duration) (bool, error) {
	repo, err := ds.GetCachedRepository(name)
	if err != nil {
		return false, err
	}

	if repo == nil {
		return true, nil // Not in cache, so it's stale
	}

	return time.Since(repo.CachedAt) > maxAge, nil
}
