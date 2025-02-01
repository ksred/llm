package validation

import (
	"os"
	"path/filepath"
	"testing"
)

// requiredDirs contains all the required directories for the project
var requiredDirs = []string{
	"cmd/example",
	"internal/ratelimit",
	"internal/auth",
	"examples/simple",
	"examples/streaming",
	"examples/advanced",
	"models/openai",
	"models/anthropic",
	"models/mock",
	"config",
	"pkg/types",
	"pkg/cost",
	".github/workflows",
}

func TestProjectStructure(t *testing.T) {
	requiredDirs := []string{
		"client",
		"config",
		"internal/validation",
		"models/openai",
		"models/anthropic",
		"pkg/cost",
		"pkg/resource",
	}

	// Get project root directory
	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// Check each required directory
	for _, dir := range requiredDirs {
		path := filepath.Join(root, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Required directory missing: %s", dir)
		}
	}
}

// findProjectRoot looks for go.mod file to determine project root
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
