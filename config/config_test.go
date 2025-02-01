package config

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ksred/llm/pkg/resource"
)

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		setupEnv  func()
		cleanEnv  func()
		wantError bool
	}{
		{
			name: "valid config with API key",
			config: &Config{
				Provider: "openai",
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
			wantError: false,
		},
		{
			name: "valid config with environment variable",
			config: &Config{
				Provider: "openai",
				Model:    "gpt-4",
			},
			setupEnv: func() {
				os.Setenv("LLM_API_KEY", "test-key")
			},
			cleanEnv: func() {
				os.Unsetenv("LLM_API_KEY")
			},
			wantError: false,
		},
		{
			name: "missing API key",
			config: &Config{
				Provider: "openai",
				Model:    "gpt-4",
			},
			wantError: true,
		},
		{
			name: "missing provider",
			config: &Config{
				APIKey:   "test-key",
				Model:    "gpt-4",
				Provider: "", // explicitly set to empty to test default
			},
			wantError: false, // changed to false since we use default provider
		},
		{
			name: "invalid provider",
			config: &Config{
				Provider: "invalid",
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
			wantError: true,
		},
		{
			name: "missing model",
			config: &Config{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanEnv != nil {
				defer tt.cleanEnv()
			}

			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Config.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	// Test default configuration
	cfg, err := NewConfig("test-key")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	// Verify defaults
	if cfg.Provider != DefaultProvider {
		t.Errorf("Default provider = %v, want %v", cfg.Provider, DefaultProvider)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Default timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}
	if cfg.MaxRetries != DefaultMaxRetries {
		t.Errorf("Default max retries = %v, want %v", cfg.MaxRetries, DefaultMaxRetries)
	}
}

func TestConfigOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		want    *Config
	}{
		{
			name: "with custom timeout",
			options: []Option{
				WithTimeout(5 * time.Second),
			},
			want: &Config{
				Timeout: 5 * time.Second,
			},
		},
		{
			name: "with custom HTTP client",
			options: []Option{
				WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
			},
			want: &Config{
				HTTPClient: &http.Client{Timeout: 10 * time.Second},
			},
		},
		{
			name: "with rate limits",
			options: []Option{
				WithRateLimit(60, 1000),
			},
			want: &Config{
				RateLimit: &RateLimit{
					RequestsPerMinute: 60,
					TokensPerMinute:   1000,
				},
			},
		},
		{
			name: "with pool config",
			options: []Option{
				WithPoolConfig(&resource.PoolConfig{
					MaxSize:       10,
					IdleTimeout:   time.Minute,
					CleanupPeriod: time.Minute,
				}),
			},
			want: &Config{
				PoolConfig: &resource.PoolConfig{
					MaxSize:       10,
					IdleTimeout:   time.Minute,
					CleanupPeriod: time.Minute,
				},
			},
		},
		{
			name: "with retry config",
			options: []Option{
				WithRetryConfig(&resource.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 100 * time.Millisecond,
					MaxInterval:     time.Second,
					Multiplier:      2.0,
				}),
			},
			want: &Config{
				RetryConfig: &resource.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 100 * time.Millisecond,
					MaxInterval:     time.Second,
					Multiplier:      2.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			for _, opt := range tt.options {
				err := opt(cfg)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(cfg, tt.want) {
				t.Errorf("got %+v, want %+v", cfg, tt.want)
			}
		})
	}
}
