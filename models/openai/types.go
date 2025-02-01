package openai

import (
	"time"

	"github.com/ksred/llm/pkg/types"
)

// openAIError represents an error response from the OpenAI API
type openAIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

// openAICompletionResponse represents a completion response from the OpenAI API
type openAICompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		LogProbs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// toResponse converts an OpenAI completion response to a generic CompletionResponse
func (r *openAICompletionResponse) toResponse() *types.CompletionResponse {
	var content string
	var finishReason string
	if len(r.Choices) > 0 {
		content = r.Choices[0].Text
		finishReason = r.Choices[0].FinishReason
	}

	return &types.CompletionResponse{
		Response: types.Response{
			ID:         r.ID,
			Created:    time.Unix(r.Created, 0),
			Provider:   "openai",
			Model:      r.Model,
			Message:    types.Message{Role: types.RoleAssistant, Content: content},
			StopReason: finishReason,
			Usage: types.Usage{
				PromptTokens:     r.Usage.PromptTokens,
				CompletionTokens: r.Usage.CompletionTokens,
				TotalTokens:      r.Usage.TotalTokens,
			},
		},
	}
}

// openAIChatResponse represents a chat completion response from the OpenAI API
type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// toResponse converts an OpenAI chat response to a generic ChatResponse
func (r *openAIChatResponse) toResponse() *types.ChatResponse {
	var message types.Message
	var finishReason string
	if len(r.Choices) > 0 {
		message = types.Message{
			Role:    types.Role(r.Choices[0].Message.Role),
			Content: r.Choices[0].Message.Content,
		}
		finishReason = r.Choices[0].FinishReason
	}

	return &types.ChatResponse{
		Response: types.Response{
			ID:         r.ID,
			Created:    time.Unix(r.Created, 0),
			Provider:   "openai",
			Model:      r.Model,
			Message:    message,
			StopReason: finishReason,
			Usage: types.Usage{
				PromptTokens:     r.Usage.PromptTokens,
				CompletionTokens: r.Usage.CompletionTokens,
				TotalTokens:      r.Usage.TotalTokens,
			},
		},
	}
}

// openAIStreamResponse represents a streaming response from the OpenAI API
type openAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

// toResponse converts an OpenAI stream response to a generic ChatResponse
func (r *openAIStreamResponse) toResponse() *types.ChatResponse {
	var message types.Message
	var finishReason string
	if len(r.Choices) > 0 {
		message = types.Message{
			Role:    types.Role(r.Choices[0].Delta.Role),
			Content: r.Choices[0].Delta.Content,
		}
		finishReason = r.Choices[0].FinishReason
	}

	return &types.ChatResponse{
		Response: types.Response{
			ID:         r.ID,
			Created:    time.Unix(r.Created, 0),
			Provider:   "openai",
			Model:      r.Model,
			Message:    message,
			StopReason: finishReason,
		},
	}
}
