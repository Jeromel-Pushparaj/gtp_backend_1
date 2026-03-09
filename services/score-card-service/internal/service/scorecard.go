package service

import (
	"fmt"
	"log"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/repository"
)

// ScoreCardService handles business logic for scorecards
type ScoreCardService struct {
	repo *repository.ScoreCardRepository
}

// NewScoreCardService creates a new scorecard service
func NewScoreCardService(repo *repository.ScoreCardRepository) *ScoreCardService {
	return &ScoreCardService{
		repo: repo,
	}
}

// CreateScoreCard creates a new scorecard
func (s *ScoreCardService) CreateScoreCard(req *models.ScoreCardRequest) (*models.ScoreCardResponse, error) {
	// Calculate the overall score
	score := req.Metrics.CalculateScore()

	// Create scorecard entity
	scoreCard := &models.ScoreCard{
		ServiceName: req.ServiceName,
		Score:       score,
		Metrics:     req.Metrics,
	}

	// Save to database
	id, err := s.repo.Create(scoreCard)
	if err != nil {
		return nil, fmt.Errorf("failed to create scorecard: %w", err)
	}

	// Build response
	response := &models.ScoreCardResponse{
		ID:          id,
		ServiceName: scoreCard.ServiceName,
		Score:       score,
		Grade:       models.GetGrade(score),
		Metrics:     scoreCard.Metrics,
		CreatedAt:   scoreCard.CreatedAt,
	}

	log.Printf("✅ Created scorecard for %s with score %.2f (Grade: %s)", 
		req.ServiceName, score, response.Grade)

	return response, nil
}

// GetScoreCard retrieves a scorecard by ID
func (s *ScoreCardService) GetScoreCard(id int64) (*models.ScoreCardResponse, error) {
	scoreCard, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get scorecard: %w", err)
	}

	response := &models.ScoreCardResponse{
		ID:          scoreCard.ID,
		ServiceName: scoreCard.ServiceName,
		Score:       scoreCard.Score,
		Grade:       models.GetGrade(scoreCard.Score),
		Metrics:     scoreCard.Metrics,
		CreatedAt:   scoreCard.CreatedAt,
	}

	return response, nil
}

// GetScoreCardsByService retrieves all scorecards for a service
func (s *ScoreCardService) GetScoreCardsByService(serviceName string) ([]*models.ScoreCardResponse, error) {
	scoreCards, err := s.repo.GetByServiceName(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get scorecards: %w", err)
	}

	responses := make([]*models.ScoreCardResponse, len(scoreCards))
	for i, sc := range scoreCards {
		responses[i] = &models.ScoreCardResponse{
			ID:          sc.ID,
			ServiceName: sc.ServiceName,
			Score:       sc.Score,
			Grade:       models.GetGrade(sc.Score),
			Metrics:     sc.Metrics,
			CreatedAt:   sc.CreatedAt,
		}
	}

	return responses, nil
}

// GetLatestScoreCard retrieves the latest scorecard for a service
func (s *ScoreCardService) GetLatestScoreCard(serviceName string) (*models.ScoreCardResponse, error) {
	scoreCard, err := s.repo.GetLatestByServiceName(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest scorecard: %w", err)
	}

	response := &models.ScoreCardResponse{
		ID:          scoreCard.ID,
		ServiceName: scoreCard.ServiceName,
		Score:       scoreCard.Score,
		Grade:       models.GetGrade(scoreCard.Score),
		Metrics:     scoreCard.Metrics,
		CreatedAt:   scoreCard.CreatedAt,
	}

	return response, nil
}

