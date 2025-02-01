package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/ksred/llm/client"
	"github.com/ksred/llm/config"
	"github.com/ksred/llm/pkg/resource"
	"github.com/ksred/llm/pkg/types"
)

func main() {
	// Load API keys from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get API keys from environment
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if openaiKey == "" || anthropicKey == "" {
		log.Fatal("Missing API keys")
	}

	// Create configurations for both providers
	providers := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "OpenAI",
			config: &config.Config{
				Provider: "openai",
				Model:    "gpt-3.5-turbo",
				APIKey:   openaiKey,
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
			},
		},
		{
			name: "Anthropic",
			config: &config.Config{
				Provider: "anthropic",
				Model:    "claude-2.1",
				APIKey:   anthropicKey,
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
			},
		},
	}

	// Test each provider
	for _, p := range providers {
		fmt.Printf("\n=== Testing %s ===\n", p.name)

		// Create client
		c, err := client.NewClient(p.config)
		if err != nil {
			log.Fatalf("Failed to create client for %s: %v", p.name, err)
		}

		// Test regular chat
		fmt.Println("\nRegular Chat:")
		resp, err := c.Chat(context.Background(), &types.ChatRequest{
			Messages: []types.Message{
				{
					Role:    types.RoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    types.RoleUser,
					Content: "What is the capital of France? Answer in one sentence.",
				},
			},
			MaxTokens: 50,
		})
		if err != nil {
			log.Printf("Chat failed for %s: %v", p.name, err)
		} else {
			fmt.Printf("Response: %s\n", resp.Message.Content)
		}

		// Test streaming chat
		fmt.Println("\nStreaming Chat:")
		stream, err := c.StreamChat(context.Background(), &types.ChatRequest{
			Messages: []types.Message{
				{
					Role:    types.RoleUser,
					Content: "Count from 1 to 5, one number per line.",
				},
			},
			MaxTokens: 50,
		})
		if err != nil {
			log.Printf("Streaming chat failed for %s: %v", p.name, err)
			continue
		}

		fmt.Print("Response: ")
		for resp := range stream {
			if resp.Error != nil {
				log.Printf("Stream error for %s: %v", p.name, resp.Error)
				break
			}
			fmt.Print(resp.Message.Content)
		}
		fmt.Println()
	}
}
