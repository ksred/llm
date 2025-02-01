package anthropic

import (
	"time"

	"github.com/ksred/llm/pkg/types"
)

// anthropicCompletionResponse represents a completion response from the Anthropic API
type anthropicCompletionResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// toResponse converts an Anthropic completion response to a generic CompletionResponse
func (r *anthropicCompletionResponse) toResponse() *types.CompletionResponse {
	content := ""
	for _, c := range r.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &types.CompletionResponse{
		Response: types.Response{
			ID:       r.ID,
			Created:  time.Now(), // Anthropic doesn't provide creation time
			Provider: "anthropic",
			Model:    r.Model,
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: content,
			},
			StopReason: r.StopReason,
			Usage: types.Usage{
				PromptTokens:     r.Usage.InputTokens,
				CompletionTokens: r.Usage.OutputTokens,
				TotalTokens:      r.Usage.InputTokens + r.Usage.OutputTokens,
			},
		},
	}
}

// anthropicStreamResponse represents a streaming response from the Anthropic API
type anthropicStreamResponse struct {
	Type    string `json:"type"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Role string `json:"role"`
}

// anthropicError represents an error response from the Anthropic API
type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Err     struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Error returns the error message
func (e *anthropicError) Error() string {
	if e.Err.Message != "" {
		return e.Err.Message
	}
	if e.Message != "" {
		return e.Message
	}
	return "unknown error"
}
