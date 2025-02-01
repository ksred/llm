package cost

import (
	"fmt"
	"sync"
	"time"

	"github.com/ksred/llm/pkg/types"
)

// TokenRates holds the cost per 1K tokens for a model
type TokenRates struct {
	PromptTokenRate      float64
	CompletionTokenRate float64
}

// UsageStats holds usage statistics for a model
type UsageStats struct {
	TotalTokens      int
	TotalCost        float64
	RequestCount     int
	AverageLatency   time.Duration
	LastRequestTime  time.Time
}

// CostTracker tracks usage and costs across providers and models
type CostTracker struct {
	mu      sync.RWMutex
	usage   map[string]map[string]*UsageStats // provider -> model -> stats
	budgets map[string]map[string]float64     // provider -> model -> budget
}

// NewCostTracker creates a new cost tracker
func NewCostTracker() *CostTracker {
	return &CostTracker{
		usage:   make(map[string]map[string]*UsageStats),
		budgets: make(map[string]map[string]float64),
	}
}

// TrackUsage records usage for a provider and model
func (c *CostTracker) TrackUsage(provider, model string, usage types.Usage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize maps if they don't exist
	if _, ok := c.usage[provider]; !ok {
		c.usage[provider] = make(map[string]*UsageStats)
	}
	if _, ok := c.usage[provider][model]; !ok {
		c.usage[provider][model] = &UsageStats{}
	}

	// Calculate cost
	rates := GetProviderRates()[provider][model]
	cost := (float64(usage.PromptTokens) * rates.PromptTokenRate / 1000) +
		(float64(usage.CompletionTokens) * rates.CompletionTokenRate / 1000)

	// Check budget if set
	if budget, ok := c.budgets[provider][model]; ok {
		currentCost := c.usage[provider][model].TotalCost
		if currentCost+cost > budget {
			return fmt.Errorf("budget exceeded for %s %s: current cost %.2f + new cost %.2f > budget %.2f",
				provider, model, currentCost, cost, budget)
		}
	}

	// Update stats
	stats := c.usage[provider][model]
	stats.TotalTokens += usage.TotalTokens
	stats.TotalCost += cost
	stats.RequestCount++
	stats.LastRequestTime = time.Now()

	return nil
}

// GetCost returns the total cost for a provider and model
func (c *CostTracker) GetCost(provider, model string) (float64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.usage[provider]; !ok {
		return 0, fmt.Errorf("no usage tracked for provider %s", provider)
	}
	if _, ok := c.usage[provider][model]; !ok {
		return 0, fmt.Errorf("no usage tracked for model %s", model)
	}

	return c.usage[provider][model].TotalCost, nil
}

// GetUsageStats returns usage statistics for a provider and model within a time range
func (c *CostTracker) GetUsageStats(provider, model string, start, end time.Time) (*UsageStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.usage[provider]; !ok {
		return nil, fmt.Errorf("no usage tracked for provider %s", provider)
	}
	if _, ok := c.usage[provider][model]; !ok {
		return nil, fmt.Errorf("no usage tracked for model %s", model)
	}

	stats := c.usage[provider][model]
	if stats.LastRequestTime.Before(start) || stats.LastRequestTime.After(end) {
		return nil, fmt.Errorf("no usage data in specified time range")
	}

	return stats, nil
}

// SetBudget sets a budget for a provider and model
func (c *CostTracker) SetBudget(provider, model string, budget float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.budgets[provider]; !ok {
		c.budgets[provider] = make(map[string]float64)
	}
	c.budgets[provider][model] = budget

	return nil
}

// GetProviderRates returns the token rates for all providers and models
func GetProviderRates() map[string]map[string]TokenRates {
	return map[string]map[string]TokenRates{
		"openai": {
			"gpt-4": {
				PromptTokenRate:      0.03,  // $0.03 per 1K tokens
				CompletionTokenRate:  0.06,  // $0.06 per 1K tokens
			},
			"gpt-3.5-turbo": {
				PromptTokenRate:      0.002, // $0.002 per 1K tokens
				CompletionTokenRate:  0.002, // $0.002 per 1K tokens
			},
		},
		"anthropic": {
			"claude-2.1": {
				PromptTokenRate:      0.008, // $0.008 per 1K tokens
				CompletionTokenRate:  0.024, // $0.024 per 1K tokens
			},
			"claude-2": {
				PromptTokenRate:      0.008, // $0.008 per 1K tokens
				CompletionTokenRate:  0.024, // $0.024 per 1K tokens
			},
			"claude-instant": {
				PromptTokenRate:      0.0008, // $0.0008 per 1K tokens
				CompletionTokenRate:  0.0024, // $0.0024 per 1K tokens
			},
		},
	}
}
