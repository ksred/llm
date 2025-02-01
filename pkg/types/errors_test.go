package types

import (
	"errors"
	"strings"
	"testing"
)

func TestProviderError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ProviderError
		expected string
	}{
		{
			name: "complete error with code",
			err: &ProviderError{
				Provider: "openai",
				Code:     "rate_limit_exceeded",
				Message:  "Too many requests",
				Err:      ErrRateLimitExceeded,
			},
			expected: "openai provider error (rate_limit_exceeded): Too many requests",
		},
		{
			name: "error without code",
			err: &ProviderError{
				Provider: "anthropic",
				Message:  "Invalid API key",
				Err:      ErrInvalidCredentials,
			},
			expected: "anthropic provider error: Invalid API key",
		},
		{
			name: "minimal error",
			err: &ProviderError{
				Provider: "mock",
				Message:  "Test error",
			},
			expected: "mock provider error: Test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ProviderError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProviderError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	provErr := &ProviderError{
		Provider: "openai",
		Code:     "internal_error",
		Message:  "Something went wrong",
		Err:      baseErr,
	}

	unwrapped := errors.Unwrap(provErr)
	if unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestNewProviderError(t *testing.T) {
	baseErr := errors.New("underlying error")
	err := NewProviderError("openai", "invalid_request", "Bad request", baseErr)

	// Type assertion
	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Fatal("NewProviderError() did not return a *ProviderError")
	}

	// Check fields
	if provErr.Provider != "openai" {
		t.Errorf("Provider = %v, want %v", provErr.Provider, "openai")
	}
	if provErr.Code != "invalid_request" {
		t.Errorf("Code = %v, want %v", provErr.Code, "invalid_request")
	}
	if provErr.Message != "Bad request" {
		t.Errorf("Message = %v, want %v", provErr.Message, "Bad request")
	}
	if provErr.Err != baseErr {
		t.Errorf("Err = %v, want %v", provErr.Err, baseErr)
	}
}

func TestCommonErrors(t *testing.T) {
	// Test that all common errors are properly defined
	commonErrors := []error{
		ErrInvalidRequest,
		ErrProviderError,
		ErrRateLimitExceeded,
		ErrContextTooLong,
		ErrInvalidCredentials,
		ErrTimeout,
	}

	for _, err := range commonErrors {
		if err == nil {
			t.Error("Common error is nil")
		}
		if strings.TrimSpace(err.Error()) == "" {
			t.Errorf("Common error %T has empty error message", err)
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		targetError error
		shouldMatch bool
	}{
		{
			name: "direct match",
			err: &ProviderError{
				Provider: "openai",
				Err:      ErrRateLimitExceeded,
			},
			targetError: ErrRateLimitExceeded,
			shouldMatch: true,
		},
		{
			name: "no match",
			err: &ProviderError{
				Provider: "openai",
				Err:      ErrRateLimitExceeded,
			},
			targetError: ErrInvalidRequest,
			shouldMatch: false,
		},
		{
			name: "nil error",
			err: &ProviderError{
				Provider: "openai",
			},
			targetError: ErrRateLimitExceeded,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := errors.Is(tt.err, tt.targetError)
			if matches != tt.shouldMatch {
				t.Errorf("errors.Is() = %v, want %v", matches, tt.shouldMatch)
			}
		})
	}
}
