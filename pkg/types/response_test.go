package types

import (
	"testing"
	"time"
)

func TestResponse_Validation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		resp      *Response
		wantError bool
	}{
		{
			name: "valid response",
			resp: &Response{
				ID:      "test-id",
				Created: now,
				Message: Message{
					Role:    RoleAssistant,
					Content: "Test response",
				},
				Provider: "test-provider",
				Model:    "test-model",
			},
			wantError: false,
		},
		{
			name: "missing ID",
			resp: &Response{
				Created: now,
				Message: Message{
					Role:    RoleAssistant,
					Content: "Test response",
				},
				Provider: "test-provider",
				Model:    "test-model",
			},
			wantError: true,
		},
		{
			name: "invalid message",
			resp: &Response{
				ID:      "test-id",
				Created: now,
				Message: Message{
					Role:    "",
					Content: "",
				},
				Provider: "test-provider",
				Model:    "test-model",
			},
			wantError: true,
		},
		{
			name: "missing provider",
			resp: &Response{
				ID:      "test-id",
				Created: now,
				Message: Message{
					Role:    RoleAssistant,
					Content: "Test response",
				},
				Model: "test-model",
			},
			wantError: true,
		},
		{
			name: "missing model",
			resp: &Response{
				ID:      "test-id",
				Created: now,
				Message: Message{
					Role:    RoleAssistant,
					Content: "Test response",
				},
				Provider: "test-provider",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Response.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestUsage_Total(t *testing.T) {
	tests := []struct {
		name     string
		usage    Usage
		expected int
	}{
		{
			name:     "zero usage",
			usage:    Usage{},
			expected: 0,
		},
		{
			name: "normal usage",
			usage: Usage{
				PromptTokens:      10,
				CompletionTokens:  20,
				TotalTokens:       30,
			},
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.usage.Total(); got != tt.expected {
				t.Errorf("Usage.Total() = %v, want %v", got, tt.expected)
			}
		})
	}
}
