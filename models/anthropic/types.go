package anthropic

import (
	"time"

	"github.com/ksred/llm/pkg/types"
)

// anthropicCompletionResponse represents a completion response from the Anthropic API
type anthropicCompletionResponse struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Role        string `json:"role"`
	Content     string `json:"content"`
	Model       string `json:"model"`
	StopReason  string `json:"stop_reason"`
	CompletedAt string `json:"completed_at"`
	Usage       struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// toResponse converts an Anthropic completion response to a generic CompletionResponse
func (r *anthropicCompletionResponse) toResponse() *types.CompletionResponse {
	return &types.CompletionResponse{
		Response: types.Response{
			ID:       r.ID,
			Created:  time.Now(), // Anthropic doesn't provide creation time
			Provider: "anthropic",
			Model:    r.Model,
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: r.Content,
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
	Content string `json:"content"`
	Role    string `json:"role"`
}

// anthropicError represents an error response from the Anthropic API
type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Error returns the error message
func (e *anthropicError) Error() string {
	return e.Message
}
