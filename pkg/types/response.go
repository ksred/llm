package types

import (
	"errors"
	"time"
)

var (
	ErrMissingID       = errors.New("response ID is required")
	ErrMissingProvider = errors.New("provider is required")
	ErrMissingModel    = errors.New("model is required")
)

// Response represents a common response structure
type Response struct {
	ID         string    `json:"id"`
	Created    time.Time `json:"created"`
	Provider   string    `json:"provider"`
	Model      string    `json:"model"`
	Message    Message   `json:"message"`
	StopReason string    `json:"stop_reason"`
	Usage      Usage     `json:"usage"`
	Error      error     `json:"-"`
}

// CompletionResponse represents a completion response
type CompletionResponse struct {
	Response
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Response
}

// Usage tracks token usage for the request and response
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIError represents an error from the LLM provider
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Validate ensures the response meets all requirements
func (r *Response) Validate() error {
	if r.ID == "" {
		return ErrMissingID
	}

	if err := r.Message.Validate(); err != nil {
		return err
	}

	if r.Provider == "" {
		return ErrMissingProvider
	}

	if r.Model == "" {
		return ErrMissingModel
	}

	return nil
}

// Total returns the total number of tokens used
func (u Usage) Total() int {
	return u.TotalTokens
}

// Error returns the error message if present
func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}
