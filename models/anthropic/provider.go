package anthropic

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
	defaultBaseURL = "https://api.anthropic.com/v1/"
	apiVersion     = "2023-06-01" // Latest stable version as of now
)

// Provider implements the Provider interface for Anthropic
type Provider struct {
	config      *config.Config
	baseURL     string
	pool        *resource.ConnectionPool
	client      *resource.RetryableClient
	retryConfig *resource.RetryConfig
}

// NewProvider creates a new Anthropic provider
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

	pool := resource.NewConnectionPool(cfg.PoolConfig, "anthropic", cfg.Metrics)
	httpClient, err := pool.Get(context.Background())
	if err != nil {
		return nil, fmt.Errorf("getting client from pool: %w", err)
	}
	client := resource.NewRetryableClient(httpClient, cfg.RetryConfig, "anthropic", cfg.Metrics)

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
	body := map[string]interface{}{
		"model":      p.config.Model,
		"prompt":     req.Prompt,
		"max_tokens": req.MaxTokens,
		"stream":     false,
	}

	var resp anthropicCompletionResponse
	if err := p.doRequest(ctx, "POST", "/complete", body, &resp); err != nil {
		return nil, err
	}

	// Convert to CompletionResponse
	content := ""
	for _, c := range resp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &types.CompletionResponse{
		Response: types.Response{
			ID:       resp.ID,
			Provider: "anthropic",
			Model:    resp.Model,
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: content,
			},
			StopReason: resp.StopReason,
			Usage: types.Usage{
				PromptTokens:     resp.Usage.InputTokens,
				CompletionTokens: resp.Usage.OutputTokens,
			},
		},
	}, nil
}

// StreamComplete streams a completion for the given prompt
func (p *Provider) StreamComplete(ctx context.Context, req *types.CompletionRequest) (<-chan *types.CompletionResponse, error) {
	body := map[string]interface{}{
		"model":      p.config.Model,
		"prompt":     req.Prompt,
		"max_tokens": req.MaxTokens,
		"stream":     true,
	}

	ch := make(chan *types.CompletionResponse)
	streamCh, err := p.streamRequest(ctx, "/complete", body)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		for resp := range streamCh {
			ch <- &types.CompletionResponse{
				Response: resp.Response,
			}
		}
	}()

	return ch, nil
}

// Chat generates a chat completion for the given messages
func (p *Provider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	// Convert messages to Anthropic format
	var systemMessage string
	userMessages := make([]map[string]string, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role == types.RoleSystem {
			systemMessage = msg.Content
			continue
		}
		userMessages = append(userMessages, map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	body := map[string]interface{}{
		"model":      p.config.Model,
		"messages":   userMessages,
		"max_tokens": req.MaxTokens,
		"stream":     false,
	}

	if systemMessage != "" {
		body["system"] = systemMessage
	}

	var resp anthropicCompletionResponse
	if err := p.doRequest(ctx, "POST", "/messages", body, &resp); err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	// Convert to ChatResponse
	content := ""
	for _, c := range resp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &types.ChatResponse{
		Response: types.Response{
			ID:       resp.ID,
			Provider: "anthropic",
			Model:    resp.Model,
			Message: types.Message{
				Role:    types.RoleAssistant,
				Content: content,
			},
			StopReason: resp.StopReason,
			Usage: types.Usage{
				PromptTokens:     resp.Usage.InputTokens,
				CompletionTokens: resp.Usage.OutputTokens,
			},
		},
	}, nil
}

// StreamChat streams a chat completion for the given messages
func (p *Provider) StreamChat(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, error) {
	// Convert messages to Anthropic format
	var systemMessage string
	userMessages := make([]map[string]string, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role == types.RoleSystem {
			systemMessage = msg.Content
			continue
		}
		userMessages = append(userMessages, map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	body := map[string]interface{}{
		"model":      p.config.Model,
		"messages":   userMessages,
		"max_tokens": req.MaxTokens,
		"stream":     true,
	}

	if systemMessage != "" {
		body["system"] = systemMessage
	}

	return p.streamRequest(ctx, "/messages", body)
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
	req.Header.Set("X-API-Key", p.config.APIKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr anthropicError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("request failed with status %d: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("provider error: %s", apiErr.Error())
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

// streamRequest handles streaming responses from the Anthropic API
func (p *Provider) streamRequest(ctx context.Context, path string, body interface{}) (<-chan *types.ChatResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.config.APIKey)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var apiErr types.ProviderError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, &apiErr
	}

	responseChan := make(chan *types.ChatResponse)

	go func() {
		defer resp.Body.Close()
		defer close(responseChan)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Remove "data: " prefix
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var streamResp anthropicStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				responseChan <- &types.ChatResponse{
					Response: types.Response{
						Error: fmt.Errorf("error decoding stream: %w", err),
					},
				}
				return
			}

			// Convert stream response to ChatResponse
			if streamResp.Type == "message_start" || streamResp.Type == "message_delta" {
				continue
			}

			if streamResp.Type == "content_block_delta" || streamResp.Type == "content_block_start" {
				content := streamResp.Delta.Text
				if content != "" {
					responseChan <- &types.ChatResponse{
						Response: types.Response{
							Message: types.Message{
								Role:    types.RoleAssistant,
								Content: content,
							},
						},
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			responseChan <- &types.ChatResponse{
				Response: types.Response{
					Error: fmt.Errorf("error reading stream: %w", err),
				},
			}
		}
	}()

	return responseChan, nil
}
