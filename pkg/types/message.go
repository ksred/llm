package types

import (
	"errors"
	"fmt"
)

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

var (
	ErrEmptyRole     = errors.New("message role cannot be empty")
	ErrInvalidRole   = errors.New("invalid message role")
	ErrEmptyContent  = errors.New("message content cannot be empty")
)

// Message represents a single message in a conversation
type Message struct {
	Role     Role         `json:"role"`
	Content  string       `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Validate ensures the message meets all requirements
func (m *Message) Validate() error {
	if m.Role == "" {
		return ErrEmptyRole
	}

	if m.Role != RoleSystem && m.Role != RoleUser && m.Role != RoleAssistant {
		return fmt.Errorf("%w: %s", ErrInvalidRole, m.Role)
	}

	if m.Content == "" {
		return ErrEmptyContent
	}

	return nil
}

// String implements the Stringer interface
func (m Message) String() string {
	return fmt.Sprintf("[%s]: %s", m.Role, m.Content)
}
