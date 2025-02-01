package examples

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/ksred/llm/client"
	"github.com/ksred/llm/config"
	"github.com/ksred/llm/pkg/resource"
	"github.com/ksred/llm/pkg/types"
)

func getCurrentFile() string {
	_, file, _, _ := runtime.Caller(0)
	return file
}

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode")
	}

	// Load .env file from project root
	projectRoot := filepath.Dir(filepath.Dir(getCurrentFile()))
	if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Get API keys from environment
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if openaiKey == "" || anthropicKey == "" {
		t.Skip("skipping end-to-end test: missing API keys")
	}

	// Test cases for different providers and configurations
	tests := []struct {
		name     string
		provider string
		model    string
		apiKey   string
	}{
		{
			name:     "OpenAI GPT-3.5",
			provider: "openai",
			model:    "gpt-3.5-turbo", // This model only works with chat completions
			apiKey:   openaiKey,
		},
		{
			name:     "Anthropic Claude",
			provider: "anthropic",
			model:    "claude-2.1", // Updated to latest Claude model
			apiKey:   anthropicKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create configuration with connection pool and retry logic
			cfg := &config.Config{
				Provider: tt.provider,
				Model:    tt.model,
				APIKey:   tt.apiKey,
				PoolConfig: &resource.PoolConfig{
					MaxSize:       5,
					IdleTimeout:   time.Minute,
					CleanupPeriod: time.Minute,
				},
				RetryConfig: &resource.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 100 * time.Millisecond,
					MaxInterval:     time.Second,
					Multiplier:      2.0,
				},
			}

			// Create client
			c, err := client.NewClient(cfg)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			// Test completion (using chat for both providers since they're chat models)
			t.Run("completion", func(t *testing.T) {
				ctx := context.Background()
				resp, err := c.Chat(ctx, &types.ChatRequest{
					Messages: []types.Message{
						{
							Role:    types.RoleUser,
							Content: "What is the capital of France?",
						},
					},
					MaxTokens: 50,
				})
				if err != nil {
					t.Fatalf("completion failed: %v", err)
				}
				if resp.Message.Content == "" {
					t.Error("empty completion response")
				}
				fmt.Printf("Completion response: %s\n", resp.Message.Content)
			})

			// Test chat
			t.Run("chat", func(t *testing.T) {
				ctx := context.Background()
				resp, err := c.Chat(ctx, &types.ChatRequest{
					Messages: []types.Message{
						{
							Role:    types.RoleSystem,
							Content: "You are a helpful assistant.",
						},
						{
							Role:    types.RoleUser,
							Content: "What is the capital of France?",
						},
					},
					MaxTokens: 50,
				})
				if err != nil {
					t.Fatalf("chat failed: %v", err)
				}
				if resp.Message.Content == "" {
					t.Error("empty chat response")
				}
				fmt.Printf("Chat response: %s\n", resp.Message.Content)
			})

			// Test streaming chat
			t.Run("streaming chat", func(t *testing.T) {
				ctx := context.Background()
				stream, err := c.StreamChat(ctx, &types.ChatRequest{
					Messages: []types.Message{
						{
							Role:    types.RoleUser,
							Content: "What is 2+2? Answer in one word.",
						},
					},
					MaxTokens: 50,
				})
				if err != nil {
					t.Fatalf("streaming chat failed: %v", err)
				}

				var fullResponse string
				for resp := range stream {
					if resp.Error != nil {
						t.Fatalf("stream error: %v", resp.Error)
					}
					fullResponse += resp.Message.Content
				}
				fmt.Printf("Streaming response: %s\n", fullResponse)
			})

			// Test error handling
			t.Run("error handling", func(t *testing.T) {
				ctx := context.Background()
				_, err := c.Chat(ctx, &types.ChatRequest{
					Messages: []types.Message{
						{
							Role:    types.RoleUser,
							Content: "What is 2+2?",
						},
					},
					MaxTokens: -1, // Invalid max tokens to trigger error
				})
				if err == nil {
					t.Error("expected error but got nil")
				}
			})

			// Test context cancellation
			t.Run("context cancellation", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
				defer cancel()

				_, err := c.Chat(ctx, &types.ChatRequest{
					Messages: []types.Message{
						{
							Role:    types.RoleUser,
							Content: "Write a very long story",
						},
					},
					MaxTokens: 1000,
				})
				if err == nil {
					t.Error("expected error but got nil")
				}
			})
		})
	}
}
