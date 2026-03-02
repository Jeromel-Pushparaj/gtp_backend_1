package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
)

// ScoreCardRepository handles database operations for scorecards
type ScoreCardRepository struct {
	db *sql.DB
}

// NewScoreCardRepository creates a new scorecard repository
func NewScoreCardRepository(db *sql.DB) *ScoreCardRepository {
	return &ScoreCardRepository{
		db: db,
	}
}

// Create inserts a new scorecard into the database
func (r *ScoreCardRepository) Create(scoreCard *models.ScoreCard) (int64, error) {
	query := `
		INSERT INTO scorecards (service_name, score, code_quality, test_coverage, 
			security_score, performance_score, documentation_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	now := time.Now()
	scoreCard.CreatedAt = now
	scoreCard.UpdatedAt = now

	var id int64
	err := r.db.QueryRow(
		query,
		scoreCard.ServiceName,
		scoreCard.Score,
		scoreCard.Metrics.CodeQuality,
		scoreCard.Metrics.TestCoverage,
		scoreCard.Metrics.SecurityScore,
		scoreCard.Metrics.PerformanceScore,
		scoreCard.Metrics.DocumentationScore,
		scoreCard.CreatedAt,
		scoreCard.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert scorecard: %w", err)
	}

	return id, nil
}

// GetByID retrieves a scorecard by ID
func (r *ScoreCardRepository) GetByID(id int64) (*models.ScoreCard, error) {
	query := `
		SELECT id, service_name, score, code_quality, test_coverage,
			security_score, performance_score, documentation_score, created_at, updated_at
		FROM scorecards
		WHERE id = $1
	`

	scoreCard := &models.ScoreCard{}
	err := r.db.QueryRow(query, id).Scan(
		&scoreCard.ID,
		&scoreCard.ServiceName,
		&scoreCard.Score,
		&scoreCard.Metrics.CodeQuality,
		&scoreCard.Metrics.TestCoverage,
		&scoreCard.Metrics.SecurityScore,
		&scoreCard.Metrics.PerformanceScore,
		&scoreCard.Metrics.DocumentationScore,
		&scoreCard.CreatedAt,
		&scoreCard.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("scorecard not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scorecard: %w", err)
	}

	return scoreCard, nil
}

// GetByServiceName retrieves all scorecards for a service
func (r *ScoreCardRepository) GetByServiceName(serviceName string) ([]*models.ScoreCard, error) {
	query := `
		SELECT id, service_name, score, code_quality, test_coverage,
			security_score, performance_score, documentation_score, created_at, updated_at
		FROM scorecards
		WHERE service_name = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to query scorecards: %w", err)
	}
	defer rows.Close()

	var scorecards []*models.ScoreCard
	for rows.Next() {
		sc := &models.ScoreCard{}
		err := rows.Scan(
			&sc.ID,
			&sc.ServiceName,
			&sc.Score,
			&sc.Metrics.CodeQuality,
			&sc.Metrics.TestCoverage,
			&sc.Metrics.SecurityScore,
			&sc.Metrics.PerformanceScore,
			&sc.Metrics.DocumentationScore,
			&sc.CreatedAt,
			&sc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scorecard: %w", err)
		}
		scorecards = append(scorecards, sc)
	}

	return scorecards, nil
}

// GetLatestByServiceName retrieves the latest scorecard for a service
func (r *ScoreCardRepository) GetLatestByServiceName(serviceName string) (*models.ScoreCard, error) {
	query := `
		SELECT id, service_name, score, code_quality, test_coverage,
			security_score, performance_score, documentation_score, created_at, updated_at
		FROM scorecards
		WHERE service_name = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	scoreCard := &models.ScoreCard{}
	err := r.db.QueryRow(query, serviceName).Scan(
		&scoreCard.ID,
		&scoreCard.ServiceName,
		&scoreCard.Score,
		&scoreCard.Metrics.CodeQuality,
		&scoreCard.Metrics.TestCoverage,
		&scoreCard.Metrics.SecurityScore,
		&scoreCard.Metrics.PerformanceScore,
		&scoreCard.Metrics.DocumentationScore,
		&scoreCard.CreatedAt,
		&scoreCard.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no scorecard found for service")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest scorecard: %w", err)
	}

	return scoreCard, nil
}

