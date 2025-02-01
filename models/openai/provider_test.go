package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"encoding/json"

	"github.com/ksred/llm/config"
	"github.com/ksred/llm/pkg/resource"
	"github.com/ksred/llm/pkg/types"
)

func TestProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "test-id",
			"choices": []map[string]interface{}{
				{
					"text": "Hello",
				},
			},
			"model": "gpt-4",
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
				Provider: "openai",
				Model:    "gpt-4",
				APIKey:   "test-key",
				BaseURL:  server.URL,
			},
			request: &types.CompletionRequest{
				Prompt: "Hello",
			},
			want: &types.CompletionResponse{
				Response: types.Response{
					ID:       "test-id",
					Provider: "openai",
					Model:    "gpt-4",
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
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "test-id",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello",
					},
				},
			},
			"model": "gpt-4",
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
				Provider: "openai",
				Model:    "gpt-4",
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
					Provider: "openai",
					Model:    "gpt-4",
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Simulate streaming response
		responses := []string{
			`{"id":"1","choices":[{"delta":{"role":"assistant","content":"Hello"}}]}`,
			`{"id":"2","choices":[{"delta":{"content":" world"}}]}`,
			`{"id":"3","choices":[{"delta":{"content":"!"}}]}`,
		}

		for _, resp := range responses {
			fmt.Fprintf(w, "data: %s\n\n", resp)
			w.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Provider: "openai",
		Model:    "gpt-4",
		APIKey:   "test-key",
		BaseURL:  server.URL,
	}

	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	stream, err := p.StreamChat(context.Background(), &types.ChatRequest{
		Messages: []types.Message{
			{
				Role:    "user",
				Content: "Hi",
			},
		},
	})

	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
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
}

func TestProvider_ConnectionPool(t *testing.T) {
	cfg := &config.Config{
		PoolConfig: &resource.PoolConfig{
			MaxSize:       2,
			IdleTimeout:   time.Second,
			CleanupPeriod: time.Second,
		},
	}

	pool := resource.NewConnectionPool(cfg.PoolConfig, "openai", nil)
	if pool == nil {
		t.Fatal("NewConnectionPool() returned nil")
	}
}

func TestProvider_RetryableClient(t *testing.T) {
	cfg := &config.Config{
		RetryConfig: &resource.RetryConfig{
			MaxRetries:      2,
			InitialInterval: time.Millisecond,
			MaxInterval:     time.Millisecond * 10,
			Multiplier:      2,
		},
	}

	pool := resource.NewConnectionPool(cfg.PoolConfig, "openai", nil)
	client, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("pool.Get() error = %v", err)
	}

	retryClient := resource.NewRetryableClient(client, cfg.RetryConfig, "openai", nil)
	if retryClient == nil {
		t.Fatal("NewRetryableClient() returned nil")
	}
}
