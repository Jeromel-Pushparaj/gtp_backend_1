package vectordb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SearchSimilarLearnings finds learnings similar to the query text
func (s *Store) SearchSimilarLearnings(ctx context.Context, queryText string, opts SearchOptions) ([]LearningResult, error) {
	// Generate embedding for query
	queryEmbedding, err := s.embeddingClient.Embed(ctx, queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.SearchSimilarLearningsByVector(ctx, queryEmbedding, opts)
}

// SearchSimilarLearningsByVector finds learnings similar to the query vector
func (s *Store) SearchSimilarLearningsByVector(ctx context.Context, queryVector Vector, opts SearchOptions) ([]LearningResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	// Use cosine distance (1 - cosine_similarity)
	// Lower distance = more similar, so we convert to score
	query := `
		SELECT 
			id, category, source_api, content, context, confidence,
			times_applied, last_applied_at, created_at, updated_at,
			1 - (embedding <=> $1) as score
		FROM learnings
		WHERE embedding IS NOT NULL
	`
	args := []interface{}{vectorToString(queryVector)}
	argIdx := 2

	if opts.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, opts.Category)
		argIdx++
	}

	if opts.APISource != "" {
		query += fmt.Sprintf(" AND source_api = $%d", argIdx)
		args = append(args, opts.APISource)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY embedding <=> $1 LIMIT $%d", argIdx)
	args = append(args, opts.Limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []LearningResult
	for rows.Next() {
		var lr LearningResult
		var contextJSON []byte
		err := rows.Scan(
			&lr.ID, &lr.Category, &lr.SourceAPI, &lr.Content, &contextJSON,
			&lr.Confidence, &lr.TimesApplied, &lr.LastApplied,
			&lr.CreatedAt, &lr.UpdatedAt, &lr.Score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if contextJSON != nil {
			json.Unmarshal(contextJSON, &lr.Context)
		}

		// Filter by minimum score
		if lr.Score >= opts.MinScore {
			results = append(results, lr)
		}
	}

	return results, nil
}

// SearchSimilarFailures finds failure patterns similar to the query
func (s *Store) SearchSimilarFailures(ctx context.Context, queryText string, opts SearchOptions) ([]FailurePatternResult, error) {
	queryEmbedding, err := s.embeddingClient.Embed(ctx, queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	query := `
		SELECT 
			id, failure_type, error_signature, error_code, endpoint_pattern,
			http_method, api_source, root_cause, fix_description, fix_payload,
			times_encountered, times_fixed, fix_success_rate,
			first_seen_at, last_seen_at, created_at, updated_at,
			1 - (embedding <=> $1) as score
		FROM failure_patterns
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, vectorToString(queryEmbedding), opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []FailurePatternResult
	for rows.Next() {
		var fr FailurePatternResult
		var fixPayloadJSON []byte
		err := rows.Scan(
			&fr.ID, &fr.FailureType, &fr.ErrorSignature, &fr.ErrorCode, &fr.EndpointPattern,
			&fr.HTTPMethod, &fr.APISource, &fr.RootCause, &fr.FixDescription, &fixPayloadJSON,
			&fr.TimesEncountered, &fr.TimesFixed, &fr.FixSuccessRate,
			&fr.FirstSeenAt, &fr.LastSeenAt, &fr.CreatedAt, &fr.UpdatedAt, &fr.Score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if fixPayloadJSON != nil {
			json.Unmarshal(fixPayloadJSON, &fr.FixPayload)
		}

		if fr.Score >= opts.MinScore {
			results = append(results, fr)
		}
	}

	return results, nil
}

// SearchSimilarStrategies finds strategies similar to the query
func (s *Store) SearchSimilarStrategies(ctx context.Context, queryText string, opts SearchOptions) ([]StrategyResult, error) {
	queryEmbedding, err := s.embeddingClient.Embed(ctx, queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	query := `
		SELECT 
			id, strategy_type, strategy_name, description, strategy_content,
			applicable_patterns, api_source, endpoint_patterns,
			success_rate, times_used, avg_coverage, created_at, updated_at,
			1 - (embedding <=> $1) as score
		FROM successful_strategies
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, vectorToString(queryEmbedding), opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []StrategyResult
	for rows.Next() {
		var sr StrategyResult
		var contentJSON, patternsJSON []byte
		err := rows.Scan(
			&sr.ID, &sr.StrategyType, &sr.StrategyName, &sr.Description, &contentJSON,
			&patternsJSON, &sr.APISource, pq.Array(&sr.EndpointPatterns),
			&sr.SuccessRate, &sr.TimesUsed, &sr.AvgCoverage, &sr.CreatedAt, &sr.UpdatedAt, &sr.Score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if contentJSON != nil {
			json.Unmarshal(contentJSON, &sr.StrategyContent)
		}
		if patternsJSON != nil {
			json.Unmarshal(patternsJSON, &sr.ApplicablePatterns)
		}

		if sr.Score >= opts.MinScore {
			results = append(results, sr)
		}
	}

	return results, nil
}

// GetLearningByID retrieves a learning by ID
func (s *Store) GetLearningByID(ctx context.Context, id uuid.UUID) (*Learning, error) {
	query := `
		SELECT id, category, source_api, content, context, confidence,
			   times_applied, last_applied_at, created_at, updated_at
		FROM learnings WHERE id = $1
	`

	var l Learning
	var contextJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.Category, &l.SourceAPI, &l.Content, &contextJSON,
		&l.Confidence, &l.TimesApplied, &l.LastApplied, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if contextJSON != nil {
		json.Unmarshal(contextJSON, &l.Context)
	}

	return &l, nil
}

// IncrementLearningUsage increments the times_applied counter for a learning
func (s *Store) IncrementLearningUsage(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE learnings 
		SET times_applied = times_applied + 1, last_applied_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// RecordFixSuccess records that a fix was successfully applied
func (s *Store) RecordFixSuccess(ctx context.Context, patternID uuid.UUID) error {
	query := `
		UPDATE failure_patterns
		SET times_fixed = times_fixed + 1,
		    fix_success_rate = times_fixed::float / times_encountered::float
		WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query, patternID)
	return err
}

