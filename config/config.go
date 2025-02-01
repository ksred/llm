package config

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/ksred/llm/pkg/resource"
	"github.com/ksred/llm/pkg/types"
)

const (
	// Environment variable names
	EnvAPIKey     = "LLM_API_KEY"
	EnvProvider   = "LLM_PROVIDER"
	EnvModel      = "LLM_MODEL"
	EnvBaseURL    = "LLM_BASE_URL"
	EnvTimeout    = "LLM_TIMEOUT"
	EnvMaxRetries = "LLM_MAX_RETRIES"

	// Default values
	DefaultProvider   = "openai"
	DefaultModel      = "gpt-4"
	DefaultTimeout    = 30 * time.Second
	DefaultMaxRetries = 3
)

var (
	// ErrMissingAPIKey is returned when no API key is provided
	ErrMissingAPIKey = errors.New("API key is required")
	// ErrMissingProvider is returned when no provider is specified
	ErrMissingProvider = errors.New("provider is required")
	// ErrInvalidProvider is returned when an unsupported provider is specified
	ErrInvalidProvider = errors.New("invalid provider")
	// ErrMissingModel is returned when no model is specified
	ErrMissingModel = errors.New("model is required")
)

// Config holds all configuration for the LLM client
type Config struct {
	// Required fields
	Provider string
	Model    string
	APIKey   string

	// Optional fields
	BaseURL     string
	HTTPClient  *http.Client
	Timeout     time.Duration
	MaxRetries  int
	RateLimit   *RateLimit
	CostControl *CostControl
	PoolConfig  *resource.PoolConfig
	RetryConfig *resource.RetryConfig
	Metrics     *types.MetricsCallbacks
}

// RateLimit defines rate limiting configuration
type RateLimit struct {
	RequestsPerMinute int
	TokensPerMinute   int
}

// CostControl defines cost control configuration
type CostControl struct {
	MaxCostPerRequest float64
	MaxCostPerDay     float64
}

// WithPoolConfig sets the connection pool configuration
func WithPoolConfig(poolConfig *resource.PoolConfig) Option {
	return func(c *Config) error {
		c.PoolConfig = poolConfig
		return nil
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(retryConfig *resource.RetryConfig) Option {
	return func(c *Config) error {
		c.RetryConfig = retryConfig
		return nil
	}
}

// Validate ensures all required fields are set
func (c *Config) Validate() error {
	// Check API key from config or environment
	if c.APIKey == "" {
		c.APIKey = os.Getenv(EnvAPIKey)
		if c.APIKey == "" {
			return ErrMissingAPIKey
		}
	}

	// Check provider from config or environment
	if c.Provider == "" {
		c.Provider = os.Getenv(EnvProvider)
		if c.Provider == "" {
			c.Provider = DefaultProvider
		}
	}

	// Validate provider
	switch c.Provider {
	case "openai", "anthropic":
		// Valid providers
	default:
		return ErrInvalidProvider
	}

	// Check model from config or environment
	if c.Model == "" {
		c.Model = os.Getenv(EnvModel)
		if c.Model == "" {
			return ErrMissingModel
		}
	}

	return nil
}

// NewConfig creates a new Config with the given API key and options
func NewConfig(apiKey string, opts ...Option) (*Config, error) {
	cfg := &Config{
		APIKey:     apiKey,
		Provider:   DefaultProvider,
		Model:      DefaultModel,
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
