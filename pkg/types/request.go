package types

import "errors"

var (
	ErrEmptyPrompt   = errors.New("prompt cannot be empty")
	ErrEmptyMessages = errors.New("messages cannot be empty")
)

// CompletionRequest represents a request for text completion
type CompletionRequest struct {
	Prompt           string         `json:"prompt"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	Temperature      float32        `json:"temperature,omitempty"`
	TopP             float32        `json:"top_p,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	PresencePenalty  float32        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32        `json:"frequency_penalty,omitempty"`
	User             string         `json:"user,omitempty"`
	RequestMetadata  map[string]any `json:"request_metadata,omitempty"`
}

// Validate ensures the completion request is valid
func (r *CompletionRequest) Validate() error {
	if r.Prompt == "" {
		return ErrEmptyPrompt
	}
	return nil
}

// ChatRequest represents a request for chat completion
type ChatRequest struct {
	Messages         []Message      `json:"messages"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	Temperature      float32        `json:"temperature,omitempty"`
	TopP             float32        `json:"top_p,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	PresencePenalty  float32        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32        `json:"frequency_penalty,omitempty"`
	User             string         `json:"user,omitempty"`
	RequestMetadata  map[string]any `json:"request_metadata,omitempty"`
}

// Validate ensures the chat request is valid
func (r *ChatRequest) Validate() error {
	if len(r.Messages) == 0 {
		return ErrEmptyMessages
	}

	for _, msg := range r.Messages {
		if err := msg.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate ensures the chat response is valid
func (r *ChatResponse) Validate() error {
	if r.ID == "" {
		return errors.New("response ID is required")
	}

	if err := r.Message.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate ensures the completion response is valid
func (r *CompletionResponse) Validate() error {
	if r.ID == "" {
		return errors.New("response ID is required")
	}

	if err := r.Message.Validate(); err != nil {
		return err
	}

	return nil
}
