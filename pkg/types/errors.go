package types

import (
	"errors"
	"fmt"
)

// Common errors that may be returned by the LLM package
var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrProviderError      = errors.New("provider error")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrContextTooLong     = errors.New("context length exceeded")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTimeout            = errors.New("request timeout")
)

// ProviderError wraps an error from an LLM provider with additional context
type ProviderError struct {
	Provider string
	Code     string
	Message  string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s provider error (%s): %s", e.Provider, e.Code, e.Message)
	}
	return fmt.Sprintf("%s provider error: %s", e.Provider, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError
func NewProviderError(provider, code, message string, err error) error {
	return &ProviderError{
		Provider: provider,
		Code:     code,
		Message:  message,
		Err:      err,
	}
}
