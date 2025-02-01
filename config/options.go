package config

import (
	"net/http"
	"time"

	"github.com/ksred/llm/pkg/types"
)

// Option is a function that modifies Config
type Option func(*Config) error

// WithProvider sets the provider
func WithProvider(provider string) Option {
	return func(c *Config) error {
		c.Provider = provider
		return nil
	}
}

// WithModel sets the model
func WithModel(model string) Option {
	return func(c *Config) error {
		c.Model = model
		return nil
	}
}

// WithBaseURL sets the base URL for API requests
func WithBaseURL(url string) Option {
	return func(c *Config) error {
		c.BaseURL = url
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) error {
		c.HTTPClient = client
		return nil
	}
}

// WithTimeout sets the timeout for requests
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = timeout
		if c.HTTPClient != nil {
			c.HTTPClient.Timeout = timeout
		}
		return nil
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(retries int) Option {
	return func(c *Config) error {
		if retries < 0 {
			retries = 0
		}
		c.MaxRetries = retries
		return nil
	}
}

// WithRateLimit sets rate limiting configuration
func WithRateLimit(requestsPerMinute, tokensPerMinute int) Option {
	return func(c *Config) error {
		c.RateLimit = &RateLimit{
			RequestsPerMinute: requestsPerMinute,
			TokensPerMinute:   tokensPerMinute,
		}
		return nil
	}
}

// WithCostControl sets cost control configuration
func WithCostControl(maxCostPerRequest, maxCostPerDay float64) Option {
	return func(c *Config) error {
		c.CostControl = &CostControl{
			MaxCostPerRequest: maxCostPerRequest,
			MaxCostPerDay:     maxCostPerDay,
		}
		return nil
	}
}

// WithMetrics sets the metrics callbacks
func WithMetrics(metrics *types.MetricsCallbacks) Option {
	return func(c *Config) error {
		c.Metrics = metrics
		return nil
	}
}
