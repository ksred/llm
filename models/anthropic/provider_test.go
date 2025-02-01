package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ksred/llm/config"
	"github.com/ksred/llm/pkg/resource"
	"github.com/ksred/llm/pkg/types"
)

func TestProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "test-id",
			"completion":  "Hello",
			"model":       "claude-2",
			"stop_reason": "stop",
		})
	}))
	defer server.Close()

	tests := []struct {
		name    string
		config  *config.Config
		request *types.CompletionRequest
		want    *types.CompletionResponse
		wantErr bool
	}{
		{
			name: "valid request",
			config: &config.Config{
				Provider: "anthropic",
				Model:    "claude-2",
				APIKey:   "test-key",
				BaseURL:  server.URL,
			},
			request: &types.CompletionRequest{
				Prompt: "Hello",
			},
			want: &types.CompletionResponse{
				Response: types.Response{
					ID:       "test-id",
					Provider: "anthropic",
					Model:    "claude-2",
					Message: types.Message{
						Role:    "assistant",
						Content: "Hello",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(tt.config)
			if err != nil {
				t.Fatalf("NewProvider() error = %v", err)
			}
			got, err := p.Complete(context.Background(), tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.Complete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Message.Role != tt.want.Message.Role {
				t.Errorf("Provider.Complete() Role = %v, want %v", got.Message.Role, tt.want.Message.Role)
			}
		})
	}
}

func TestProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "test-id",
			"completion":  "Hello",
			"model":       "claude-2",
			"stop_reason": "stop",
		})
	}))
	defer server.Close()

	tests := []struct {
		name    string
		config  *config.Config
		request *types.ChatRequest
		want    *types.ChatResponse
		wantErr bool
	}{
		{
			name: "valid chat request",
			config: &config.Config{
				Provider: "anthropic",
				Model:    "claude-2",
				APIKey:   "test-key",
				BaseURL:  server.URL,
			},
			request: &types.ChatRequest{
				Messages: []types.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			want: &types.ChatResponse{
				Response: types.Response{
					ID:       "test-id",
					Provider: "anthropic",
					Model:    "claude-2",
					Message: types.Message{
						Role:    "assistant",
						Content: "Hello",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(tt.config)
			if err != nil {
				t.Fatalf("NewProvider() error = %v", err)
			}
			got, err := p.Chat(context.Background(), tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.Chat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Message.Role != tt.want.Message.Role {
				t.Errorf("Provider.Chat() Role = %v, want %v", got.Message.Role, tt.want.Message.Role)
			}
		})
	}
}

func TestProvider_StreamChat(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		request *types.ChatRequest
		wantErr bool
	}{
		{
			name: "valid stream request",
			config: &config.Config{
				Provider: "anthropic",
				Model:    "claude-2",
				APIKey:   "test-key",
			},
			request: &types.ChatRequest{
				Messages: []types.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-API-Key") != "test-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)

				responses := []map[string]interface{}{
					{
						"type":    "content",
						"content": "Hello",
					},
					{
						"type":    "content",
						"content": " world",
					},
					{
						"type":    "content",
						"content": "!",
					},
				}

				for _, resp := range responses {
					data, _ := json.Marshal(resp)
					fmt.Fprintf(w, "%s\n", data)
					w.(http.Flusher).Flush()
					time.Sleep(10 * time.Millisecond)
				}
			}))
			defer server.Close()

			tt.config.BaseURL = server.URL
			p, err := NewProvider(tt.config)
			if err != nil {
				t.Fatalf("NewProvider() error = %v", err)
			}

			stream, err := p.StreamChat(context.Background(), tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.StreamChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			var messages []string
			for resp := range stream {
				if resp.Error != nil {
					t.Errorf("StreamChat() error in response: %v", resp.Error)
					continue
				}
				messages = append(messages, resp.Message.Content)
			}

			want := []string{"Hello", " world", "!"}
			if !reflect.DeepEqual(messages, want) {
				t.Errorf("StreamChat() got messages = %v, want %v", messages, want)
			}
		})
	}
}

func TestProvider_ConnectionPool(t *testing.T) {
	cfg := &config.Config{
		PoolConfig: &resource.PoolConfig{
			MaxSize:       2,
			IdleTimeout:   time.Second,
			CleanupPeriod: time.Second,
		},
	}

	metrics := &types.MetricsCallbacks{
		OnPoolGet: func(provider string, waitTime time.Duration) {
			// No-op for testing
		},
		OnPoolRelease: func(provider string) {
			// No-op for testing
		},
		OnPoolExhausted: func(provider string) {
			// No-op for testing
		},
	}

	pool := resource.NewConnectionPool(cfg.PoolConfig, "anthropic", metrics)
	if pool == nil {
		t.Fatal("NewConnectionPool() returned nil")
	}

	// Get a client
	client, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Put it back
	pool.Put(client)

	// Should be able to get it again
	client2, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if client2 == nil {
		t.Fatal("Get() returned nil client")
	}
}

func TestProvider_RetryableClient(t *testing.T) {
	cfg := &config.Config{
		PoolConfig: &resource.PoolConfig{
			MaxSize:       2,
			IdleTimeout:   time.Second,
			CleanupPeriod: time.Second,
		},
	}

	metrics := &types.MetricsCallbacks{
		OnPoolGet: func(provider string, waitTime time.Duration) {
			// No-op for testing
		},
		OnPoolRelease: func(provider string) {
			// No-op for testing
		},
		OnPoolExhausted: func(provider string) {
			// No-op for testing
		},
	}

	pool := resource.NewConnectionPool(cfg.PoolConfig, "anthropic", metrics)
	if pool == nil {
		t.Fatal("NewConnectionPool() returned nil")
	}

	client, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("pool.Get() error = %v", err)
	}

	retryClient := resource.NewRetryableClient(client, &resource.RetryConfig{
		MaxRetries:      2,
		InitialInterval: time.Millisecond,
		MaxInterval:     time.Millisecond * 10,
		Multiplier:      2,
	}, "anthropic", nil)
	if retryClient == nil {
		t.Fatal("NewRetryableClient() returned nil")
	}
}

func TestProvider_RetryableClient_Retry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{
		Provider: "anthropic",
		Model:    "claude-2",
		APIKey:   "test-key",
		BaseURL:  server.URL,
		RetryConfig: &resource.RetryConfig{
			MaxRetries:      2,
			InitialInterval: time.Millisecond,
			MaxInterval:     10 * time.Millisecond,
			Multiplier:      2,
		},
	}

	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	req := &types.CompletionRequest{
		Prompt: "Hello",
	}

	_, err = p.Complete(context.Background(), req)
	if err == nil {
		t.Error("Complete() expected error after max retries")
	}
}
