package cost

import (
	"testing"
	"time"

	"github.com/ksred/llm/pkg/types"
)

func TestCostTracker_TrackUsage(t *testing.T) {
	tracker := NewCostTracker()

	usage := types.Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	// Track usage for OpenAI GPT-4
	err := tracker.TrackUsage("openai", "gpt-4", usage)
	if err != nil {
		t.Errorf("TrackUsage() error = %v", err)
	}

	// Get cost for the tracked usage
	cost, err := tracker.GetCost("openai", "gpt-4")
	if err != nil {
		t.Errorf("GetCost() error = %v", err)
	}

	// GPT-4 costs $0.03 per 1K prompt tokens and $0.06 per 1K completion tokens
	expectedCost := 0.003 + 0.003 // (100 * 0.03 / 1000) + (50 * 0.06 / 1000)
	if cost != expectedCost {
		t.Errorf("GetCost() = %v, want %v", cost, expectedCost)
	}
}

func TestCostTracker_GetUsageStats(t *testing.T) {
	tracker := NewCostTracker()

	// Track usage for multiple requests
	usages := []types.Usage{
		{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
		{PromptTokens: 200, CompletionTokens: 100, TotalTokens: 300},
	}

	for _, usage := range usages {
		err := tracker.TrackUsage("openai", "gpt-4", usage)
		if err != nil {
			t.Errorf("TrackUsage() error = %v", err)
		}
	}

	// Get usage stats
	stats, err := tracker.GetUsageStats("openai", "gpt-4", time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		t.Errorf("GetUsageStats() error = %v", err)
	}

	expectedTotalTokens := 450 // 150 + 300
	if stats.TotalTokens != expectedTotalTokens {
		t.Errorf("GetUsageStats().TotalTokens = %v, want %v", stats.TotalTokens, expectedTotalTokens)
	}
}

func TestCostTracker_CheckBudget(t *testing.T) {
	tracker := NewCostTracker()

	// Set a budget of $1.00
	err := tracker.SetBudget("openai", "gpt-4", 1.00)
	if err != nil {
		t.Errorf("SetBudget() error = %v", err)
	}

	// Track usage that should exceed budget
	usage := types.Usage{
		PromptTokens:     20000, // $0.60 at $0.03 per 1K tokens
		CompletionTokens: 10000, // $0.60 at $0.06 per 1K tokens
	}

	// First usage should be rejected as it exceeds budget
	err = tracker.TrackUsage("openai", "gpt-4", usage)
	if err == nil {
		t.Error("TrackUsage() should return error when usage would exceed budget")
	}

	// Try with smaller usage that fits within budget
	smallerUsage := types.Usage{
		PromptTokens:     10000, // $0.30 at $0.03 per 1K tokens
		CompletionTokens: 5000,  // $0.30 at $0.06 per 1K tokens
	}

	err = tracker.TrackUsage("openai", "gpt-4", smallerUsage)
	if err != nil {
		t.Errorf("TrackUsage() error = %v", err)
	}

	// Verify cost
	cost, err := tracker.GetCost("openai", "gpt-4")
	if err != nil {
		t.Errorf("GetCost() error = %v", err)
	}

	expectedCost := 0.60 // $0.30 + $0.30
	if cost != expectedCost {
		t.Errorf("GetCost() = %v, want %v", cost, expectedCost)
	}
}

func TestCostTracker_GetProviderRates(t *testing.T) {
	rates := GetProviderRates()

	// Check OpenAI GPT-4 rates
	gpt4Rates, ok := rates["openai"]["gpt-4"]
	if !ok {
		t.Error("GetProviderRates() should include rates for GPT-4")
	}

	expectedPromptRate := 0.03
	if gpt4Rates.PromptTokenRate != expectedPromptRate {
		t.Errorf("GPT-4 prompt rate = %v, want %v", gpt4Rates.PromptTokenRate, expectedPromptRate)
	}

	expectedCompletionRate := 0.06
	if gpt4Rates.CompletionTokenRate != expectedCompletionRate {
		t.Errorf("GPT-4 completion rate = %v, want %v", gpt4Rates.CompletionTokenRate, expectedCompletionRate)
	}
}
