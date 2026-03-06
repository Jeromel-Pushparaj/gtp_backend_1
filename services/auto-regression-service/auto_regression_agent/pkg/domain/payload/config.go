package payload

import "time"

// MutationConfig configures the mutation generator
type MutationConfig struct {
	EnableFieldMutation bool
	EnableTypeMutation  bool
	EnableBoundaryTests bool
	CombinationDepth    int
	MutationStrategies  []MutationStrategy
}

// MutationStrategy represents a mutation strategy
type MutationStrategy struct {
	Type        string
	TargetTypes []string
	Priority    int
}

// DefaultMutationConfig returns a default mutation configuration
func DefaultMutationConfig() *MutationConfig {
	return &MutationConfig{
		EnableFieldMutation: true,
		EnableTypeMutation:  true,
		EnableBoundaryTests: true,
		CombinationDepth:    2,
		MutationStrategies: []MutationStrategy{
			{Type: "null", TargetTypes: []string{"string", "integer", "object", "array"}, Priority: 8},
			{Type: "empty", TargetTypes: []string{"string", "array", "object"}, Priority: 7},
		},
	}
}

// GeneratedPayload represents a generated test payload
type GeneratedPayload struct {
	Name           string
	Description    string
	Category       string
	Payload        interface{}
	Expected       int
	ExpectedStatus int
	TestType       string
	MutationType   string
	TargetField    string
}

// PerformanceTestConfig configures performance test generation
type PerformanceTestConfig struct {
	EnableConcurrentRequests bool
	EnableLoadRampUp         bool
	EnableSpikeTest          bool
	EnableEnduranceTest      bool
	ConcurrencyLevels        []int
	RampUpDuration           time.Duration
	SpikeDuration            time.Duration
	EnduranceDuration        time.Duration
	MaxRequestsPerSecond     int
}

// PerformanceTest represents a performance test scenario
type PerformanceTest struct {
	Name             string
	Description      string
	TestType         string
	ConcurrencyLevel int
	Duration         time.Duration
	RampUpTime       time.Duration
	ExpectedRPS      int
	SLALatency       time.Duration
}

// SecurityTestConfig configures security test generation
type SecurityTestConfig struct {
	EnableSQLInjection     bool
	EnableXSS              bool
	EnableCommandInjection bool
	EnablePathTraversal    bool
	EnableXMLInjection     bool
	EnableLDAPInjection    bool
	EnableAuthBypass       bool
	OWASPCategories        []string
}

// SecurityTest represents a security test scenario
type SecurityTest struct {
	Name            string
	Description     string
	Category        string
	Payload         interface{}
	ExpectedStatus  int
	OWASPCategory   string
	CVEReferences   []string
	SecurityImpact  string
	RemediationTips string
}

// SmartDataConfig configures smart data generation
type SmartDataConfig struct {
	UseRealisticData    bool
	LocalePreference    string
	DomainSpecificRules map[string]string
	DataConstraints     map[string]DataConstraint
	HistoricalPatterns  map[string]HistoricalPattern
}

// DataConstraint represents a constraint on generated data
type DataConstraint struct {
	MinLength int
	MaxLength int
	MinValue  float64
	MaxValue  float64
	Pattern   string
	Format    string
	Enum      []interface{}
}

// HistoricalPattern represents a pattern learned from historical data
type HistoricalPattern struct {
	FieldName     string
	CommonValues  []interface{}
	Distribution  map[string]float64
	SuccessValues []interface{}
	FailureValues []interface{}
}
