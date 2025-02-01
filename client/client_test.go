package client

import (
	"context"
	"testing"
	"time"

	"github.com/ksred/llm/config"
	"github.com/ksred/llm/pkg/types"
)

type mockProvider struct{}

func (m *mockProvider) Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	return &types.CompletionResponse{
		Response: types.Response{
			ID: "test-id",
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: "Test response",
			},
		},
	}, nil
}

func (m *mockProvider) StreamComplete(ctx context.Context, req *types.CompletionRequest) (<-chan *types.CompletionResponse, error) {
	ch := make(chan *types.CompletionResponse)
	go func() {
		defer close(ch)
		ch <- &types.CompletionResponse{
			Response: types.Response{
				ID: "test-id",
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "Test response",
				},
			},
		}
	}()
	return ch, nil
}

func (m *mockProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{
		Response: types.Response{
			ID: "test-id",
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: "Test response",
			},
		},
	}, nil
}

func (m *mockProvider) StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error) {
	ch := make(chan *types.ChatResponse)
	go func() {
		defer close(ch)
		responses := []*types.ChatResponse{
			{
				Response: types.Response{
					ID: "test-id-1",
					Message: types.Message{
						Role:    types.RoleAssistant,
						Content: "Hello",
					},
				},
			},
			{
				Response: types.Response{
					ID: "test-id-2",
					Message: types.Message{
						Role:    types.RoleAssistant,
						Content: " world!",
					},
				},
			},
		}
		for _, resp := range responses {
			select {
			case <-ctx.Done():
				return
			case ch <- resp:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
	return ch, nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		wantError bool
	}{
		{
			name: "valid configuration",
			cfg: &config.Config{
				Provider: "mock",
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
			wantError: false,
		},
		{
			name:      "missing configuration",
			cfg:       nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantError {
				t.Errorf("NewClient() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_Complete(t *testing.T) {
	client := &Client{
		config: &config.Config{
			Provider: "mock",
			APIKey:   "test-key",
			Model:    "test-model",
		},
		provider: &mockProvider{},
	}

	resp, err := client.Complete(context.Background(), &types.CompletionRequest{
		Prompt: "Test prompt",
	})

	if err != nil {
		t.Errorf("Complete() error = %v", err)
		return
	}

	expected := &types.CompletionResponse{
		Response: types.Response{
			ID: "test-id",
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: "Test response",
			},
		},
	}

	if resp.ID != expected.ID {
		t.Errorf("Complete() got ID = %v, want %v", resp.ID, expected.ID)
	}
	if resp.Message.Role != expected.Message.Role {
		t.Errorf("Complete() got Message.Role = %v, want %v", resp.Message.Role, expected.Message.Role)
	}
	if resp.Message.Content != expected.Message.Content {
		t.Errorf("Complete() got Message.Content = %v, want %v", resp.Message.Content, expected.Message.Content)
	}
}

func TestClient_StreamChat(t *testing.T) {
	client := &Client{
		config: &config.Config{
			Provider: "mock",
			APIKey:   "test-key",
			Model:    "test-model",
		},
		provider: &mockProvider{},
	}

	stream, err := client.StreamChat(context.Background(), &types.ChatRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Hello",
			},
		},
	})

	if err != nil {
		t.Errorf("StreamChat() error = %v", err)
		return
	}

	expected := []*types.ChatResponse{
		{
			Response: types.Response{
				ID: "test-id-1",
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: "Hello",
				},
			},
		},
		{
			Response: types.Response{
				ID: "test-id-2",
				Message: types.Message{
					Role:    types.RoleAssistant,
					Content: " world!",
				},
			},
		},
	}

	i := 0
	for resp := range stream {
		if resp.Error != nil {
			t.Errorf("StreamChat() received error: %v", resp.Error)
			continue
		}

		if i >= len(expected) {
			t.Errorf("StreamChat() received more responses than expected")
			break
		}

		if resp.ID != expected[i].ID {
			t.Errorf("StreamChat() got ID = %v, want %v", resp.ID, expected[i].ID)
		}
		if resp.Message.Role != expected[i].Message.Role {
			t.Errorf("StreamChat() got Message.Role = %v, want %v", resp.Message.Role, expected[i].Message.Role)
		}
		if resp.Message.Content != expected[i].Message.Content {
			t.Errorf("StreamChat() got Message.Content = %v, want %v", resp.Message.Content, expected[i].Message.Content)
		}

		i++
	}

	if i != len(expected) {
		t.Errorf("StreamChat() received %d responses, want %d", i, len(expected))
	}
}
