package resource

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ksred/llm/pkg/types" // Assuming types package is in your-project/types
)

// PoolConfig holds configuration for the connection pool
type PoolConfig struct {
	MaxSize       int           // Maximum number of connections
	IdleTimeout   time.Duration // How long to keep idle connections
	CleanupPeriod time.Duration // How often to clean up idle connections
}

// ConnectionPool manages a pool of http.Client connections
type ConnectionPool struct {
	config   *PoolConfig
	provider string
	metrics  *types.MetricsCallbacks
	idle     []*http.Client
	active   map[*http.Client]time.Time
	mu       sync.Mutex
	shutdown bool
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *PoolConfig, provider string, metrics *types.MetricsCallbacks) *ConnectionPool {
	if config == nil {
		config = &PoolConfig{
			MaxSize:       10,
			IdleTimeout:   time.Minute,
			CleanupPeriod: time.Minute,
		}
	}
	pool := &ConnectionPool{
		config:   config,
		provider: provider,
		metrics:  metrics,
		idle:     make([]*http.Client, 0),
		active:   make(map[*http.Client]time.Time),
	}
	go pool.cleanup()
	return pool
}

// Get retrieves a client from the pool or creates a new one
func (p *ConnectionPool) Get(ctx context.Context) (*http.Client, error) {
	start := time.Now()
	for {
		p.mu.Lock()
		if p.shutdown {
			p.mu.Unlock()
			return nil, fmt.Errorf("pool is shut down")
		}

		// Try to get an idle client
		if len(p.idle) > 0 {
			client := p.idle[len(p.idle)-1]
			p.idle = p.idle[:len(p.idle)-1]
			p.active[client] = time.Now()
			p.mu.Unlock()

			if p.metrics != nil && p.metrics.OnPoolGet != nil {
				p.metrics.OnPoolGet(p.provider, time.Since(start))
			}
			return client, nil
		}

		// Check if we can create a new client
		if len(p.active) < p.config.MaxSize {
			// Create new client
			client := &http.Client{
				Timeout: 30 * time.Second,
			}
			p.active[client] = time.Now()
			p.mu.Unlock()

			if p.metrics != nil && p.metrics.OnPoolGet != nil {
				p.metrics.OnPoolGet(p.provider, time.Since(start))
			}
			return client, nil
		}

		// Pool is exhausted
		if p.metrics != nil && p.metrics.OnPoolExhausted != nil {
			p.metrics.OnPoolExhausted(p.provider)
		}

		p.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Try again
		}
	}
}

// Put returns a client to the pool
func (p *ConnectionPool) Put(client *http.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return
	}

	delete(p.active, client)
	p.idle = append(p.idle, client)

	if p.metrics != nil && p.metrics.OnPoolRelease != nil {
		p.metrics.OnPoolRelease(p.provider)
	}
}

// cleanup periodically removes idle connections
func (p *ConnectionPool) cleanup() {
	ticker := time.NewTicker(p.config.CleanupPeriod)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		if p.shutdown {
			p.mu.Unlock()
			return
		}

		now := time.Now()
		remaining := make([]*http.Client, 0, len(p.idle))

		// Remove idle clients that have timed out
		for _, client := range p.idle {
			if lastUsed, ok := p.active[client]; ok {
				if now.Sub(lastUsed) < p.config.IdleTimeout {
					remaining = append(remaining, client)
				}
			}
		}

		p.idle = remaining
		p.mu.Unlock()
	}
}

// Shutdown closes the pool and all connections
func (p *ConnectionPool) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.shutdown = true
	p.idle = nil
	p.active = nil
	return nil
}

// RetryConfig configures the retry behavior
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

// NewRetryableClient creates a new retryable client
func NewRetryableClient(client *http.Client, config *RetryConfig, provider string, metrics *types.MetricsCallbacks) *RetryableClient {
	if config == nil {
		config = &RetryConfig{
			MaxRetries:      3,
			InitialInterval: time.Second,
			MaxInterval:     30 * time.Second,
			Multiplier:      2,
		}
	}
	return &RetryableClient{
		client:   client,
		config:   config,
		provider: provider,
		metrics:  metrics,
	}
}

// RetryableClient wraps an http.Client with retry logic
type RetryableClient struct {
	client   *http.Client
	config   *RetryConfig
	provider string
	metrics  *types.MetricsCallbacks
}

// Do executes an HTTP request with retries
func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	interval := c.config.InitialInterval

	start := time.Now()
	if c.metrics != nil && c.metrics.OnRequest != nil {
		c.metrics.OnRequest(c.provider)
	}

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Sleep before retry with exponential backoff
			time.Sleep(interval)
			interval = time.Duration(float64(interval) * c.config.Multiplier)
			if interval > c.config.MaxInterval {
				interval = c.config.MaxInterval
			}

			if c.metrics != nil && c.metrics.OnRetry != nil {
				c.metrics.OnRetry(c.provider, attempt, err)
			}
		}

		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			if c.metrics != nil && c.metrics.OnResponse != nil {
				c.metrics.OnResponse(c.provider, time.Since(start))
			}
			return resp, nil
		}

		if err != nil && c.metrics != nil && c.metrics.OnError != nil {
			c.metrics.OnError(c.provider, err)
		}

		// Close the response body if we're going to retry
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	// If we've exhausted all retries, return an error
	if resp != nil && resp.StatusCode >= 500 {
		err = fmt.Errorf("server error: %d", resp.StatusCode)
	} else if err == nil {
		err = fmt.Errorf("max retries exceeded")
	}

	return nil, err
}
