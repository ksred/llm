# Go LLM Package Technical Specification

## Project Overview
A production-ready Go package for interacting with Large Language Models (LLMs), supporting multiple providers, streaming, cost management, and robust error handling.

## Project Structure

```
llm/
├── cmd/
│   └── example/                  # Example CLI implementation
│       └── main.go              # Example usage of the package
├── internal/
│   ├── tokenizer/
│   │   ├── tokenizer.go         # Token counting interface
│   │   ├── tiktoken.go          # TikToken implementation
│   │   └── tokenizer_test.go
│   ├── ratelimit/
│   │   ├── ratelimit.go         # Rate limiting implementation
│   │   └── ratelimit_test.go
│   └── auth/
│       ├── auth.go              # Authentication handling
│       └── auth_test.go
├── examples/
│   ├── simple/                  # Basic usage examples
│   ├── streaming/               # Streaming examples
│   └── advanced/               # Advanced configuration examples
├── models/
│   ├── types.go                # Shared model interfaces
│   ├── openai/
│   │   ├── client.go
│   │   ├── chat.go
│   │   ├── completions.go
│   │   └── openai_test.go
│   ├── anthropic/
│   │   ├── client.go
│   │   ├── messages.go
│   │   └── anthropic_test.go
│   └── mock/
│       ├── mock.go             # Mock model for testing
│       └── mock_test.go
├── config/
│   ├── config.go               # Configuration interface
│   ├── options.go              # Configuration options
│   ├── validation.go           # Configuration validation
│   └── config_test.go
├── pkg/
│   ├── types/
│   │   ├── message.go          # Message types
│   │   ├── response.go         # Response types
│   │   └── errors.go           # Error types
│   └── cost/
│       ├── tracker.go          # Cost tracking
│       └── calculator.go       # Cost calculation
├── .github/
│   ├── workflows/
│   │   ├── test.yml           # CI testing
│   │   └── release.yml        # Release automation
│   └── ISSUE_TEMPLATE/
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Core Interfaces

### Client Implementation
```go
// ClientOptions holds all client configuration
type ClientOptions struct {
    Provider    string
    APIKey      string
    Model       string
    BaseURL     string
    HTTPClient  *http.Client
    Timeout     time.Duration
    RetryConfig *RetryConfig
    RateLimit   *RateLimitConfig
    CostConfig  *CostConfig
    Logger      Logger
}

// Option is a function that modifies ClientOptions
type Option func(*ClientOptions) error

// Client is the main LLM client interface
type Client interface {
    // Complete generates a completion for the given prompt
    Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error)
    
    // StreamComplete streams a completion for the given prompt
    StreamComplete(ctx context.Context, req *types.CompletionRequest) (<-chan *types.CompletionResponse, error)
    
    // Chat generates a chat completion
    Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error)
    
    // StreamChat streams a chat completion
    StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error)
    
    // GetUsage returns current usage statistics
    GetUsage() *types.Usage
}
```

### Client Configuration

The client uses a functional options pattern for configuration. The API key is a required parameter, while other configurations are optional.

```go
// Constructor requires API key
func NewClient(apiKey string, opts ...Option) (*Client, error) {
    if apiKey == "" {
        return nil, ErrMissingAPIKey
    }
    
    // Default options
    options := &ClientOptions{
        Provider: "openai",  // Default provider
        APIKey: apiKey,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
    
    // Apply options
    for _, opt := range opts {
        if err := opt(options); err != nil {
            return nil, fmt.Errorf("applying option: %w", err)
        }
    }
    
    // Validate configuration
    if err := options.validate(); err != nil {
        return nil, err
    }
    
    return &Client{
        options: options,
        // Initialize other fields
    }, nil
}

// Provider-specific constructors
func NewOpenAIClient(apiKey string, opts ...Option) (*Client, error) {
    return NewClient(
        apiKey,
        append(opts, WithProvider("openai"))...,
    )
}

func NewAnthropicClient(apiKey string, opts ...Option) (*Client, error) {
    return NewClient(
        apiKey,
        append(opts, WithProvider("anthropic"))...,
    )
}
    
    // Request configuration
    GetTimeout() time.Duration
    GetMaxRetries() int
    GetBackoffConfig() BackoffConfig
    
    // Rate limiting
    GetRateLimits() RateLimitConfig
    
    // Cost management
    GetCostConfig() CostConfig
    
    // Logger configuration
    GetLogger() Logger
}

type BackoffConfig struct {
    InitialDuration time.Duration
    MaxDuration     time.Duration
    Multiplier      float64
}

type RateLimitConfig struct {
    RequestsPerMinute int
    TokensPerMinute   int
    ConcurrentCalls   int
}

type CostConfig struct {
    MaxCostPerRequest float64
    BudgetPeriod     time.Duration
    MaxBudget        float64
}
```

### Message Types
```go
type Message struct {
    Role     string          `json:"role"`
    Content  string          `json:"content"`
    Metadata map[string]any  `json:"metadata,omitempty"`
}

type CompletionRequest struct {
    Prompt          string            `json:"prompt"`
    MaxTokens       int               `json:"max_tokens,omitempty"`
    Temperature     float32           `json:"temperature,omitempty"`
    TopP            float32           `json:"top_p,omitempty"`
    Stop            []string          `json:"stop,omitempty"`
    PresencePenalty float32          `json:"presence_penalty,omitempty"`
    FrequencyPenalty float32         `json:"frequency_penalty,omitempty"`
    Options         map[string]any    `json:"options,omitempty"`
}

type ChatRequest struct {
    Messages        []Message         `json:"messages"`
    MaxTokens       int               `json:"max_tokens,omitempty"`
    Temperature     float32           `json:"temperature,omitempty"`
    TopP            float32           `json:"top_p,omitempty"`
    Stop            []string          `json:"stop,omitempty"`
    PresencePenalty float32          `json:"presence_penalty,omitempty"`
    FrequencyPenalty float32         `json:"frequency_penalty,omitempty"`
    Options         map[string]any    `json:"options,omitempty"`
}

type CompletionResponse struct {
    ID              string           `json:"id"`
    Text            string           `json:"text"`
    FinishReason    string           `json:"finish_reason"`
    Usage           Usage            `json:"usage"`
    Created         time.Time        `json:"created"`
    Model           string           `json:"model"`
    SystemFingerprint string         `json:"system_fingerprint,omitempty"`
}

type ChatResponse struct {
    ID              string           `json:"id"`
    Messages        []Message        `json:"messages"`
    FinishReason    string           `json:"finish_reason"`
    Usage           Usage            `json:"usage"`
    Created         time.Time        `json:"created"`
    Model           string           `json:"model"`
    SystemFingerprint string         `json:"system_fingerprint,omitempty"`
}

type Usage struct {
    PromptTokens     int     `json:"prompt_tokens"`
    CompletionTokens int     `json:"completion_tokens"`
    TotalTokens      int     `json:"total_tokens"`
    Cost            float64 `json:"cost"`
}
```

## Error Handling

### Custom Error Types
```go
type Error struct {
    Code    ErrorCode
    Message string
    Err     error
}

type ErrorCode int

const (
    ErrConfiguration ErrorCode = iota + 1
    ErrAuthentication
    ErrRateLimit
    ErrTimeout
    ErrCanceled
    ErrInvalidRequest
    ErrProviderError
    ErrCostExceeded
)
```

## Configuration

### Configuration Options

```go
// Standard options
func WithProvider(provider string) Option {
    return func(o *ClientOptions) error {
        o.Provider = provider
        return nil
    }
}

func WithModel(model string) Option {
    return func(o *ClientOptions) error {
        o.Model = model
        return nil
    }
}

func WithBaseURL(url string) Option {
    return func(o *ClientOptions) error {
        o.BaseURL = url
        return nil
    }
}

func WithHTTPClient(client *http.Client) Option {
    return func(o *ClientOptions) error {
        o.HTTPClient = client
        return nil
    }
}

func WithTimeout(timeout time.Duration) Option {
    return func(o *ClientOptions) error {
        o.Timeout = timeout
        return nil
    }
}

func WithRetryConfig(config RetryConfig) Option {
    return func(o *ClientOptions) error {
        o.RetryConfig = &config
        return nil
    }
}

func WithRateLimit(config RateLimitConfig) Option {
    return func(o *ClientOptions) error {
        o.RateLimit = &config
        return nil
    }
}

func WithCostConfig(config CostConfig) Option {
    return func(o *ClientOptions) error {
        o.CostConfig = &config
        return nil
    }
}

func WithLogger(logger Logger) Option {
    return func(o *ClientOptions) error {
        o.Logger = logger
        return nil
    }
}
```

### Environment Variables
```
# Required
LLM_API_KEY=sk-...

# Optional with defaults
LLM_PROVIDER=openai
LLM_MODEL=gpt-4
LLM_TIMEOUT=30s
LLM_MAX_RETRIES=3
LLM_REQUESTS_PER_MINUTE=60
LLM_MAX_TOKENS_PER_MINUTE=100000
LLM_MAX_CONCURRENT_CALLS=10
LLM_MAX_COST_PER_REQUEST=0.50
LLM_MAX_BUDGET=100.00
LLM_BUDGET_PERIOD=24h
LLM_LOG_LEVEL=info
```

### YAML Configuration
```yaml
provider: openai
model: gpt-4
timeout: 30s
retries:
  max_attempts: 3
  initial_delay: 100ms
  max_delay: 2s
  multiplier: 2.0
rate_limits:
  requests_per_minute: 60
  tokens_per_minute: 100000
  concurrent_calls: 10
cost_control:
  max_cost_per_request: 0.50
  max_budget: 100.00
  budget_period: 24h
logging:
  level: info
  format: json
```

## Integration Examples

The examples directory provides practical integration patterns for different use cases.

### Simple Usage (examples/simple/main.go)
```go
package main

import (
    "context"
    "log"
    "os"
    "time"
    
    "github.com/yourusername/llm"
    "github.com/yourusername/llm/pkg/types"
)

func main() {
    // Create client with direct configuration
    client, err := llm.NewClient(
        "sk-your-api-key",
        llm.WithProvider("openai"),
        llm.WithModel("gpt-4"),
        llm.WithTimeout(30 * time.Second),
    )
    
    // Or create client from environment variables
    client, err := llm.NewClient(
        os.Getenv("LLM_API_KEY"),
        llm.WithProvider(os.Getenv("LLM_PROVIDER")),
        llm.WithModel(os.Getenv("LLM_MODEL")),
    )
    
    // Or use provider-specific constructor
    client, err := llm.NewOpenAIClient(
        os.Getenv("OPENAI_API_KEY"),
        llm.WithModel("gpt-4"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Create completion request
    req := &types.CompletionRequest{
        Prompt:    "Translate this to French: Hello, world!",
        MaxTokens: 100,
    }
    
    // Get completion
    resp, err := client.Complete(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", resp.Text)
    log.Printf("Usage: %+v", resp.Usage)
}
```

### Streaming Integration (examples/streaming/main.go)
```go
func main() {
    client, err := llm.NewClient(
        llm.WithModel("gpt-4"),
        llm.WithTimeout(30 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    req := &types.ChatRequest{
        Messages: []types.Message{
            {
                Role:    "user",
                Content: "Write a story about a brave knight.",
            },
        },
        MaxTokens: 1000,
    }
    
    stream, err := client.StreamChat(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    for resp := range stream {
        if resp.Error != nil {
            log.Printf("Error: %v", resp.Error)
            continue
        }
        fmt.Print(resp.Messages[0].Content)
    }
}
```

## Testing Strategy

### Unit Tests
- Every package should have comprehensive unit tests
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for >80% code coverage

### Integration Tests
- Test interaction between components
- Test configuration loading
- Test rate limiting behavior
- Test cost tracking accuracy

### Mock Provider
- Implement mock provider for testing
- Simulate various error conditions
- Test retry behavior
- Test streaming behavior

## Documentation

### Package Documentation
- Comprehensive godoc comments
- Example functions for common use cases
- Clear error documentation
- Configuration guide

### README
- Quick start guide
- Installation instructions
- Configuration examples
- Common use cases
- Contributing guidelines

## Implementation Guidelines

### Error Handling
- Use custom error types
- Include context in errors
- Wrap underlying errors
- Provide clear error messages

### Concurrency
- Use context for cancellation
- Implement proper resource cleanup
- Handle goroutine leaks
- Use sync primitives correctly

### Rate Limiting
- Use token bucket algorithm
- Implement per-model limits
- Handle concurrent requests
- Provide backpressure

### Retry Logic
- Implement exponential backoff
- Handle specific error types
- Respect context cancellation
- Track retry attempts

## Cost Management

### Token Counting
- Implement accurate token counting
- Support different tokenizers
- Cache token counts
- Handle streaming token counts

### Cost Tracking
- Track cost per request
- Implement budget limits
- Provide usage reports
- Support cost alerts

## Security Considerations

### API Key Management
- Secure key storage
- Key rotation support
- Mask keys in logs
- Environment isolation

### Request/Response Security
- Sanitize logging
- Handle sensitive data
- Implement timeouts
- Validate inputs

## Performance Optimization

### Connection Management
- Connection pooling
- Keep-alive settings
- Request timeouts
- Circuit breaking

### Memory Management
- Buffer pooling
- Stream processing
- Garbage collection
- Resource limits

## Deployment Considerations

### Version Management
- Semantic versioning
- Changelog maintenance
- Deprecation policy
- Migration guides

### CI/CD
- Automated testing
- Code coverage
- Linting
- Release automation

## Contributing Guidelines

### Code Style
- Follow Go standards
- Use gofmt
- Implement golint
- Document changes

### Advanced Configuration (examples/advanced/main.go)
```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/yourusername/llm"
    "github.com/yourusername/llm/pkg/types"
)

func main() {
    // Create client with advanced configuration
    client, err := llm.NewClient(
        os.Getenv("OPENAI_API_KEY"),
        llm.WithProvider("openai"),
        llm.WithModel("gpt-4"),
        llm.WithTimeout(30 * time.Second),
        llm.WithRetryConfig(llm.RetryConfig{
            MaxAttempts: 3,
            InitialDelay: 100 * time.Millisecond,
            MaxDelay: 2 * time.Second,
            Multiplier: 2.0,
        }),
        llm.WithRateLimit(llm.RateLimitConfig{
            RequestsPerMinute: 60,
            TokensPerMinute: 100000,
            ConcurrentCalls: 10,
        }),
        llm.WithCostConfig(llm.CostConfig{
            MaxCostPerRequest: 0.50,
            MaxBudget: 100.00,
            BudgetPeriod: 24 * time.Hour,
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use the client with advanced features
    req := &types.ChatRequest{
        Messages: []types.Message{
            {
                Role:    "system",
                Content: "You are a helpful assistant.",
            },
            {
                Role:    "user",
                Content: "Write a story about a brave knight.",
            },
        },
        MaxTokens: 1000,
    }

    // Get response with context timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    resp, err := client.Chat(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    // Access usage statistics
    log.Printf("Response: %s", resp.Messages[len(resp.Messages)-1].Content)
    log.Printf("Token usage: %d tokens", resp.Usage.TotalTokens)
    log.Printf("Cost: $%.4f", resp.Usage.Cost)
}
```

### Web Framework Integration (examples/integration/fiber/main.go)
```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/yourusername/llm"
    "github.com/yourusername/llm/pkg/types"
)

func main() {
    app := fiber.New()
    
    // Initialize LLM client
    client, err := llm.NewOpenAIClient(
        os.Getenv("OPENAI_API_KEY"),
        llm.WithModel("gpt-4"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Chat completion endpoint
    app.Post("/chat", func(c *fiber.Ctx) error {
        var req types.ChatRequest
        if err := c.BodyParser(&req); err != nil {
            return err
        }

        resp, err := client.Chat(c.Context(), &req)
        if err != nil {
            return err
        }

        return c.JSON(resp)
    })

    // Streaming chat completion endpoint
    app.Post("/chat/stream", func(c *fiber.Ctx) error {
        var req types.ChatRequest
        if err := c.BodyParser(&req); err != nil {
            return err
        }

        c.Set("Content-Type", "text/event-stream")
        c.Set("Cache-Control", "no-cache")
        c.Set("Connection", "keep-alive")

        stream, err := client.StreamChat(c.Context(), &req)
        if err != nil {
            return err
        }

        for resp := range stream {
            if resp.Error != nil {
                return resp.Error
            }
            
            if err := c.WriteString(resp.Messages[0].Content); err != nil {
                return err
            }
            c.WriteString("\n")
            
            if c.Context().Done() != nil {
                return nil
            }
        }

        return nil
    })

    app.Listen(":3000")
}

This specification provides a comprehensive framework for implementing a production-ready LLM client package in Go. Implementation should follow Go best practices and idioms while maintaining flexibility for future enhancements.