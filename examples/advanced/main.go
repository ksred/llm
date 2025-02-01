package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/ksred/llm/client"
	"github.com/ksred/llm/config"
	"github.com/ksred/llm/pkg/cost"
	"github.com/ksred/llm/pkg/resource"
	"github.com/ksred/llm/pkg/types"
)

// Metrics tracks various performance metrics for each provider
type Metrics struct {
	totalRequests     int
	totalInputTokens  int
	totalOutputTokens int
	totalLatency      time.Duration
	successfulCalls   int
	failedCalls       int
	avgResponseTime   time.Duration
	mu                sync.Mutex
}

func (m *Metrics) recordRequest(duration time.Duration, inputTokens, outputTokens int, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	m.totalLatency += duration
	m.totalInputTokens += inputTokens
	m.totalOutputTokens += outputTokens

	if success {
		m.successfulCalls++
	} else {
		m.failedCalls++
	}
	m.avgResponseTime = m.totalLatency / time.Duration(m.totalRequests)
}

func (m *Metrics) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return fmt.Sprintf(
		"Requests: %d (Success: %d, Failed: %d)\n"+
			"Average Response Time: %v\n"+
			"Total Tokens: %d (Input: %d, Output: %d)\n",
		m.totalRequests, m.successfulCalls, m.failedCalls,
		m.avgResponseTime,
		m.totalInputTokens+m.totalOutputTokens, m.totalInputTokens, m.totalOutputTokens,
	)
}

// Provider represents an LLM provider with its client and conversation history
type Provider struct {
	name        string
	provider    string
	model       string
	client      *client.Client
	history     []types.Message
	metrics     *Metrics
	lastUsed    time.Time
	color       color.Attribute
	costTracker *cost.CostTracker
}

// newProvider creates a new provider with the given configuration
func newProvider(name, providerType, model, apiKey string, color color.Attribute) (*Provider, error) {
	cfg := &config.Config{
		Provider: providerType,
		Model:    model,
		APIKey:   apiKey,
		PoolConfig: &resource.PoolConfig{
			MaxSize:       10,
			IdleTimeout:   5 * time.Minute,
			CleanupPeriod: time.Minute,
		},
		RetryConfig: &resource.RetryConfig{
			MaxRetries:      5,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     2 * time.Second,
			Multiplier:      2.0,
		},
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating client for %s: %w", name, err)
	}

	return &Provider{
		name:        name,
		provider:    providerType,
		model:       model,
		client:      c,
		history:     make([]types.Message, 0),
		metrics:     &Metrics{},
		color:       color,
		costTracker: cost.NewCostTracker(),
	}, nil
}

// chat sends a message to the provider and returns the response
func (p *Provider) chat(ctx context.Context, msg string, stream bool) (string, error) {
	start := time.Now()
	var response string

	// Add user message to history
	p.history = append(p.history, types.Message{
		Role:    types.RoleUser,
		Content: msg,
	})

	// Prepare request with conversation history
	req := &types.ChatRequest{
		Messages:  p.history,
		MaxTokens: 1000,
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if stream {
		// Handle streaming response
		streamChan, err := p.client.StreamChat(ctx, req)
		if err != nil {
			p.metrics.recordRequest(time.Since(start), 0, 0, false)
			return "", fmt.Errorf("starting stream: %w", err)
		}

		var sb strings.Builder
		fmt.Print(color.New(p.color).Sprint("\n" + p.name + ":\n"))
		for resp := range streamChan {
			select {
			case <-ctx.Done():
				p.metrics.recordRequest(time.Since(start), 0, 0, false)
				return sb.String(), ctx.Err()
			default:
				if resp.Error != nil {
					p.metrics.recordRequest(time.Since(start), 0, 0, false)
					return sb.String(), fmt.Errorf("stream error: %w", resp.Error)
				}
				content := resp.Message.Content
				if content != "" {
					sb.WriteString(content)
					fmt.Print(color.New(p.color).Sprint(content))
				}
			}
		}
		response = sb.String()
	} else {
		// Handle regular response
		resp, err := p.client.Chat(ctx, req)
		if err != nil {
			p.metrics.recordRequest(time.Since(start), 0, 0, false)
			return "", fmt.Errorf("chat error: %w", err)
		}
		response = resp.Message.Content
	}

	// Only add assistant response to history if we got a complete response
	if response != "" {
		p.history = append(p.history, types.Message{
			Role:    types.RoleAssistant,
			Content: response,
		})
	}

	// Track usage and metrics
	inputTokens := len(msg)
	outputTokens := len(response)
	p.metrics.recordRequest(time.Since(start), inputTokens, outputTokens, true)

	// Track cost
	usage := types.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		TotalTokens:      inputTokens + outputTokens,
	}
	if err := p.costTracker.TrackUsage(p.provider, p.model, usage); err != nil {
		log.Printf("Error tracking cost: %v", err)
	}

	p.lastUsed = time.Now()
	return response, nil
}

func main() {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create cleanup function
	cleanup := func() {
		fmt.Println("\nShutting down gracefully...")
		cancel()
		// Give ongoing requests a chance to complete
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	// Handle signals in a separate goroutine
	go func() {
		<-sigChan
		cleanup()
	}()

	// Ensure cleanup runs on normal exit
	defer cleanup()

	// Load API keys
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create providers
	providers := make(map[string]*Provider)
	providerConfigs := []struct {
		name   string
		typ    string
		model  string
		apiKey string
		envKey string
		color  color.Attribute
	}{
		{
			name:   "GPT-3.5",
			typ:    "openai",
			model:  "gpt-3.5-turbo",
			envKey: "OPENAI_API_KEY",
			color:  color.FgGreen,
		},
		{
			name:   "Claude",
			typ:    "anthropic",
			model:  "claude-2.1",
			envKey: "ANTHROPIC_API_KEY",
			color:  color.FgBlue,
		},
	}

	// Initialize providers
	for _, cfg := range providerConfigs {
		apiKey := os.Getenv(cfg.envKey)
		if apiKey == "" {
			log.Printf("Warning: %s API key not found, skipping", cfg.name)
			continue
		}

		p, err := newProvider(cfg.name, cfg.typ, cfg.model, apiKey, cfg.color)
		if err != nil {
			log.Printf("Error creating %s provider: %v", cfg.name, err)
			continue
		}
		providers[cfg.name] = p
	}

	if len(providers) == 0 {
		log.Fatal("No providers available")
	}

	// Print welcome message
	fmt.Println("=== Advanced LLM Client Example ===")
	fmt.Println("Available commands:")
	fmt.Println("  /help     - Show this help message")
	fmt.Println("  /metrics  - Show provider metrics")
	fmt.Println("  /clear    - Clear conversation history")
	fmt.Println("  /quit     - Exit the program")
	fmt.Println("Type your message and press Enter to chat.")
	fmt.Println("Messages will be sent to all available providers.")
	fmt.Println("=============================")

	// Start interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("\nYou: ")
			if !scanner.Scan() {
				return
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Handle commands
			switch input {
			case "/help":
				fmt.Println("Available commands:")
				fmt.Println("  /help     - Show this help message")
				fmt.Println("  /metrics  - Show provider metrics")
				fmt.Println("  /clear    - Clear conversation history")
				fmt.Println("  /quit     - Exit the program")
				continue
			case "/metrics":
				for name, p := range providers {
					fmt.Printf("\n=== %s Metrics ===\n", name)
					fmt.Println(p.metrics)

					// Print cost metrics
					cost, err := p.costTracker.GetCost(p.provider, p.model)
					if err != nil {
						fmt.Printf("Error getting cost: %v\n", err)
					} else {
						fmt.Printf("Total Cost: $%.4f\n", cost)
					}
				}
				continue
			case "/clear":
				for _, p := range providers {
					p.history = make([]types.Message, 0)
				}
				fmt.Println("Conversation history cleared.")
				continue
			case "/quit":
				cleanup()
				return
			}

			// Send message to all providers in parallel
			var wg sync.WaitGroup
			for _, p := range providers {
				wg.Add(1)
				go func(p *Provider) {
					defer wg.Done()

					if _, err := p.chat(ctx, input, true); err != nil {
						fmt.Printf("\nError from %s: %v\n", p.name, err)
					}
				}(p)
			}
			wg.Wait()
		}
	}
}
