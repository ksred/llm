package resource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConnectionPool_Get(t *testing.T) {
	cfg := &PoolConfig{
		MaxSize:       2,
		IdleTimeout:   time.Second,
		CleanupPeriod: time.Second,
	}

	pool := NewConnectionPool(cfg, "test", nil)
	if pool == nil {
		t.Fatal("NewConnectionPool() returned nil")
	}

	// Get first client
	client1, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if client1 == nil {
		t.Fatal("Get() returned nil client")
	}

	// Get second client
	client2, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if client2 == nil {
		t.Fatal("Get() returned nil client")
	}

	// Third get should block until timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = pool.Get(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Get() error = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestConnectionPool_Put(t *testing.T) {
	cfg := &PoolConfig{
		MaxSize:       2,
		IdleTimeout:   time.Second,
		CleanupPeriod: time.Second,
	}

	pool := NewConnectionPool(cfg, "test", nil)
	if pool == nil {
		t.Fatal("NewConnectionPool() returned nil")
	}

	client, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	pool.Put(client)

	// Should be able to get the same client back
	client2, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if client2 != client {
		t.Error("Get() returned different client")
	}
}

func TestConnectionPool_Cleanup(t *testing.T) {
	cfg := &PoolConfig{
		MaxSize:       2,
		IdleTimeout:   100 * time.Millisecond,
		CleanupPeriod: 50 * time.Millisecond,
	}

	pool := NewConnectionPool(cfg, "test", nil)
	if pool == nil {
		t.Fatal("NewConnectionPool() returned nil")
	}

	client, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	pool.Put(client)

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Should get a new client
	client2, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if client2 == client {
		t.Error("Get() returned same client after cleanup")
	}
}

func TestRetryableClient_Do(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		responses  []int
		wantErr    bool
	}{
		{
			name:       "success on first try",
			maxRetries: 3,
			responses:  []int{http.StatusOK},
			wantErr:    false,
		},
		{
			name:       "success after retry",
			maxRetries: 3,
			responses:  []int{http.StatusInternalServerError, http.StatusOK},
			wantErr:    false,
		},
		{
			name:       "max retries exceeded",
			maxRetries: 2,
			responses:  []int{http.StatusInternalServerError, http.StatusInternalServerError, http.StatusInternalServerError},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if responseIndex >= len(tt.responses) {
					t.Fatalf("More requests than expected responses")
				}
				w.WriteHeader(tt.responses[responseIndex])
				responseIndex++
			}))
			defer server.Close()

			client := &http.Client{}
			retryClient := NewRetryableClient(client, &RetryConfig{
				MaxRetries:      tt.maxRetries,
				InitialInterval: time.Millisecond,
				MaxInterval:     10 * time.Millisecond,
				Multiplier:      2,
			}, "test", nil)

			req, _ := http.NewRequest("GET", server.URL, nil)
			resp, err := retryClient.Do(req)

			if tt.wantErr {
				if err == nil {
					t.Error("Do() expected error but got nil")
				}
				if resp != nil && resp.StatusCode == http.StatusOK {
					t.Error("Do() got success status code when error was expected")
				}
			} else {
				if err != nil {
					t.Errorf("Do() unexpected error: %v", err)
				} else if resp.StatusCode != http.StatusOK {
					t.Errorf("Do() got status %d, want %d", resp.StatusCode, http.StatusOK)
				}
			}
		})
	}
}

// mockHTTPClient implements http.RoundTripper for testing
type mockHTTPClient struct {
	responses []int
	current   int
	onRequest func()
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.onRequest != nil {
		m.onRequest()
	}

	status := http.StatusInternalServerError
	if m.current < len(m.responses) {
		status = m.responses[m.current]
		m.current++
	}

	return &http.Response{
		StatusCode: status,
		Body:       http.NoBody,
	}, nil
}
