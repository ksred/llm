package client

import (
	"context"
	"fmt"

	"github.com/ksred/llm/config"
	"github.com/ksred/llm/models/anthropic"
	"github.com/ksred/llm/models/openai"
	"github.com/ksred/llm/pkg/types"
)

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error)

	// StreamComplete streams a completion for the given prompt
	StreamComplete(ctx context.Context, req *types.CompletionRequest) (<-chan *types.CompletionResponse, error)

	// Chat generates a chat completion for the given messages
	Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error)

	// StreamChat streams a chat completion for the given messages
	StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error)
}

// Client is the main LLM client that delegates to specific providers
type Client struct {
	config   *config.Config
	provider Provider
}

// NewClient creates a new LLM client with the given configuration
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	// Create provider based on configuration
	var provider Provider
	switch cfg.Provider {
	case "openai":
		p, err := openai.NewProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("creating OpenAI provider: %w", err)
		}
		provider = p
	case "anthropic":
		p, err := anthropic.NewProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("creating Anthropic provider: %w", err)
		}
		provider = p
	case "mock":
		return &Client{
			config: cfg,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	return &Client{
		config:   cfg,
		provider: provider,
	}, nil
}

// Complete generates a completion for the given prompt
func (c *Client) Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	if err := c.validateRequest(ctx); err != nil {
		return nil, err
	}

	return c.provider.Complete(ctx, req)
}

// StreamComplete streams a completion for the given prompt
func (c *Client) StreamComplete(ctx context.Context, req *types.CompletionRequest) (<-chan *types.CompletionResponse, error) {
	if err := c.validateRequest(ctx); err != nil {
		return nil, err
	}

	return c.provider.StreamComplete(ctx, req)
}

// Chat generates a chat completion for the given messages
func (c *Client) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	if err := c.validateRequest(ctx); err != nil {
		return nil, err
	}

	return c.provider.Chat(ctx, req)
}

// StreamChat streams a chat completion for the given messages
func (c *Client) StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error) {
	if err := c.validateRequest(ctx); err != nil {
		return nil, err
	}

	return c.provider.StreamChat(ctx, req)
}

// validateRequest performs common validation for all requests
func (c *Client) validateRequest(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
