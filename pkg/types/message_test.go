package types

import (
	"encoding/json"
	"testing"
)

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr bool
	}{
		{
			name: "valid user message",
			message: Message{
				Role:    RoleUser,
				Content: "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "valid assistant message",
			message: Message{
				Role:    RoleAssistant,
				Content: "Hello! How can I help you?",
			},
			wantErr: false,
		},
		{
			name: "valid system message",
			message: Message{
				Role:    RoleSystem,
				Content: "You are a helpful assistant.",
			},
			wantErr: false,
		},
		{
			name: "empty role",
			message: Message{
				Content: "Hello!",
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			message: Message{
				Role:    "invalid",
				Content: "Hello!",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			message: Message{
				Role: RoleUser,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_JSON(t *testing.T) {
	msg := Message{
		Role:     RoleUser,
		Content:  "Hello!",
		Metadata: map[string]any{"timestamp": "2025-02-01T08:10:53Z"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if decoded.Role != msg.Role {
		t.Errorf("Role mismatch: got %v, want %v", decoded.Role, msg.Role)
	}
	if decoded.Content != msg.Content {
		t.Errorf("Content mismatch: got %v, want %v", decoded.Content, msg.Content)
	}
}
