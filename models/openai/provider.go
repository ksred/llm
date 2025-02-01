package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ksred/llm/config"
	"github.com/ksred/llm/pkg/resource"
	"github.com/ksred/llm/pkg/types"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	completionPath = "/completions"
	chatPath       = "/chat/completions"
)

// Provider implements the LLM provider interface for OpenAI
type Provider struct {
	config      *config.Config
	baseURL     string
	pool        *resource.ConnectionPool
	client      *resource.RetryableClient
	retryConfig *resource.RetryConfig
}

// NewProvider creates a new OpenAI provider
func NewProvider(cfg *config.Config) (*Provider, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	if cfg.PoolConfig == nil {
		cfg.PoolConfig = &resource.PoolConfig{
			MaxSize:       10,
			IdleTimeout:   time.Minute,
			CleanupPeriod: time.Minute,
		}
	}

	pool := resource.NewConnectionPool(cfg.PoolConfig, "openai", cfg.Metrics)
	httpClient, err := pool.Get(context.Background())
	if err != nil {
		return nil, fmt.Errorf("getting client from pool: %w", err)
	}
	client := resource.NewRetryableClient(httpClient, cfg.RetryConfig, "openai", cfg.Metrics)

	retryConfig := cfg.RetryConfig
	if retryConfig == nil {
		retryConfig = &resource.RetryConfig{
			MaxRetries:      3,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     time.Second,
			Multiplier:      2.0,
		}
	}

	return &Provider{
		config:      cfg,
		baseURL:     baseURL,
		pool:        pool,
		client:      client,
		retryConfig: retryConfig,
	}, nil
}

// Complete generates a completion for the given prompt
func (p *Provider) Complete(ctx context.Context, req *types.CompletionRequest) (*types.CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":             p.config.Model,
		"prompt":            req.Prompt,
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"stop":              req.Stop,
		"presence_penalty":  req.PresencePenalty,
		"frequency_penalty": req.FrequencyPenalty,
		"user":              req.User,
	}

	var resp openAICompletionResponse
	if err := p.doRequest(ctx, "POST", completionPath, body, &resp); err != nil {
		return nil, err
	}

	return resp.toResponse(), nil
}

// StreamComplete streams a completion for the given prompt
func (p *Provider) StreamComplete(ctx context.Context, req *types.CompletionRequest) (<-chan *types.CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":             p.config.Model,
		"prompt":            req.Prompt,
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"stop":              req.Stop,
		"presence_penalty":  req.PresencePenalty,
		"frequency_penalty": req.FrequencyPenalty,
		"user":              req.User,
		"stream":            true,
	}

	responseChan := make(chan *types.CompletionResponse)
	go func() {
		defer close(responseChan)

		streamChan, err := p.streamRequest(ctx, completionPath, body)
		if err != nil {
			responseChan <- &types.CompletionResponse{
				Response: types.Response{
					Error: err,
				},
			}
			return
		}

		for resp := range streamChan {
			if resp.Error != nil {
				responseChan <- &types.CompletionResponse{
					Response: types.Response{
						Error: resp.Error,
					},
				}
				continue
			}

			responseChan <- &types.CompletionResponse{
				Response: resp.Response,
			}
		}
	}()

	return responseChan, nil
}

// Chat generates a chat completion for the given messages
func (p *Provider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":             p.config.Model,
		"messages":          req.Messages,
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"stop":              req.Stop,
		"presence_penalty":  req.PresencePenalty,
		"frequency_penalty": req.FrequencyPenalty,
		"user":              req.User,
	}

	var resp openAIChatResponse
	if err := p.doRequest(ctx, "POST", chatPath, body, &resp); err != nil {
		return nil, err
	}

	return resp.toResponse(), nil
}

// StreamChat streams a chat completion for the given messages
func (p *Provider) StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":             p.config.Model,
		"messages":          req.Messages,
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"stop":              req.Stop,
		"presence_penalty":  req.PresencePenalty,
		"frequency_penalty": req.FrequencyPenalty,
		"user":              req.User,
		"stream":            true,
	}

	return p.streamRequest(ctx, chatPath, body)
}

func (p *Provider) doRequest(ctx context.Context, method, path string, body interface{}, v interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr types.ProviderError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return &apiErr
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// streamRequest handles streaming responses from the OpenAI API
func (p *Provider) streamRequest(ctx context.Context, path string, body interface{}) (<-chan *types.ChatResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var errResp openAIError
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("decoding error response: %w", err)
		}
		return nil, types.NewProviderError("openai", errResp.Error.Type, errResp.Error.Message, nil)
	}

	responseChan := make(chan *types.ChatResponse)
	go func() {
		defer resp.Body.Close()
		defer close(responseChan)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					responseChan <- &types.ChatResponse{
						Response: types.Response{
							Error: fmt.Errorf("reading stream: %w", err),
						},
					}
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || line == "data: [DONE]" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			var streamResp openAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				responseChan <- &types.ChatResponse{
					Response: types.Response{
						Error: fmt.Errorf("decoding stream response: %w", err),
					},
				}
				continue
			}

			response := streamResp.toResponse()
			select {
			case <-ctx.Done():
				return
			case responseChan <- response:
			}
		}
	}()

	return responseChan, nil
}
