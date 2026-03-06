-- OpenTest Vector Database Schema
-- This migration creates tables for storing learned patterns, failure patterns,
-- and successful strategies with vector embeddings for similarity search.

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Learnings table: Store learned patterns (auth patterns, edge cases, business rules)
CREATE TABLE IF NOT EXISTS learnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Content and context
    category VARCHAR(100) NOT NULL,  -- 'auth_pattern', 'edge_case', 'business_rule', 'api_pattern'
    source_api VARCHAR(255),          -- API/spec this was learned from
    content TEXT NOT NULL,            -- The actual learning (description)
    context JSONB,                    -- Additional context (endpoint, method, etc.)
    
    -- Vector embedding for similarity search (Gemini text-embedding-004 = 768 dimensions)
    embedding vector(768),
    
    -- Metadata
    confidence FLOAT DEFAULT 0.5,     -- How confident we are in this learning (0-1)
    times_applied INT DEFAULT 0,      -- How many times this learning was successfully applied
    last_applied_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Failure patterns table: Store failure signatures and fixes
CREATE TABLE IF NOT EXISTS failure_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Failure identification
    failure_type VARCHAR(100) NOT NULL,  -- 'auth_failure', 'validation_error', 'timeout', 'schema_mismatch'
    error_signature TEXT NOT NULL,        -- The error message/pattern that identifies this failure
    error_code VARCHAR(50),               -- HTTP status code or error code
    
    -- Context about where failure occurred
    endpoint_pattern VARCHAR(500),        -- Endpoint pattern (e.g., '/users/{id}')
    http_method VARCHAR(10),              -- GET, POST, PUT, DELETE, etc.
    api_source VARCHAR(255),              -- Which API/spec this came from
    
    -- Solution
    root_cause TEXT,                      -- Identified root cause
    fix_description TEXT NOT NULL,        -- How to fix/avoid this failure
    fix_payload JSONB,                    -- Example fixed payload if applicable
    
    -- Vector embedding for similarity search
    embedding vector(768),
    
    -- Metadata
    times_encountered INT DEFAULT 1,      -- How many times we've seen this
    times_fixed INT DEFAULT 0,            -- How many times the fix worked
    fix_success_rate FLOAT DEFAULT 0.0,   -- Success rate of the fix
    
    -- Timestamps
    first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Successful strategies table: Store test strategies that worked well
CREATE TABLE IF NOT EXISTS successful_strategies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Strategy identification
    strategy_type VARCHAR(100) NOT NULL,  -- 'auth_flow', 'data_generation', 'test_sequence', 'edge_case_coverage'
    strategy_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    
    -- Strategy details
    strategy_content JSONB NOT NULL,      -- The actual strategy (steps, payloads, etc.)
    applicable_patterns JSONB,            -- What patterns this strategy works for
    
    -- Context
    api_source VARCHAR(255),              -- API/spec this was used with
    endpoint_patterns TEXT[],             -- Endpoints this worked for
    
    -- Vector embedding for similarity search
    embedding vector(768),
    
    -- Effectiveness metrics
    success_rate FLOAT DEFAULT 0.0,       -- How often this strategy succeeds
    times_used INT DEFAULT 0,             -- How many times we've used it
    avg_coverage FLOAT DEFAULT 0.0,       -- Average test coverage achieved
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance

-- Learnings indexes
CREATE INDEX IF NOT EXISTS idx_learnings_category ON learnings(category);
CREATE INDEX IF NOT EXISTS idx_learnings_source_api ON learnings(source_api);
CREATE INDEX IF NOT EXISTS idx_learnings_embedding ON learnings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Failure patterns indexes
CREATE INDEX IF NOT EXISTS idx_failure_patterns_type ON failure_patterns(failure_type);
CREATE INDEX IF NOT EXISTS idx_failure_patterns_error_code ON failure_patterns(error_code);
CREATE INDEX IF NOT EXISTS idx_failure_patterns_embedding ON failure_patterns USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Successful strategies indexes
CREATE INDEX IF NOT EXISTS idx_strategies_type ON successful_strategies(strategy_type);
CREATE INDEX IF NOT EXISTS idx_strategies_embedding ON successful_strategies USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for auto-updating updated_at
CREATE TRIGGER update_learnings_updated_at
    BEFORE UPDATE ON learnings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_failure_patterns_updated_at
    BEFORE UPDATE ON failure_patterns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_strategies_updated_at
    BEFORE UPDATE ON successful_strategies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

