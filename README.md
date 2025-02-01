# Go LLM Client 🤖

A robust, production-ready Go client for interacting with Large Language Models. Currently supports OpenAI and Anthropic providers with a unified interface, advanced features, and enterprise-grade reliability.

[![Go Reference](https://pkg.go.dev/badge/github.com/ksred/llm.svg)](https://pkg.go.dev/github.com/ksred/llm)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksred/llm)](https://goreportcard.com/report/github.com/ksred/llm)

## Features 🌟

### Core Features
- 🔄 Unified interface for multiple LLM providers
- 📡 Real-time response streaming
- 💾 Conversation history management
- 📊 Performance metrics and cost tracking
- 🛡️ Robust error handling and retries
- 🔧 Highly configurable and extensible
- 🔌 Provider interface for easy integration of new LLMs

### Supported Providers
- OpenAI (GPT-3.5, GPT-4)
- Anthropic (Claude-2.1, Claude-2, Claude-instant)

### Coming Soon 🔜
- Mistral AI (Mistral-7B, Mixtral)
- Grok-1
- DeepSeek
- Custom Provider Interface (Bring Your Own LLM)

### Enterprise Features
- 🏊‍♂️ Connection pooling for efficient resource usage
- ⚡ Automatic retries with exponential backoff
- 💰 Cost tracking and budget management
- 🔍 Detailed usage analytics
- 🛑 Graceful error handling and recovery
- 🔒 Thread-safe operations

## Installation 📦

```bash
go get github.com/ksred/llm
```

## Quick Start 🚀

```go
package main

import (
    "context"
    "fmt"
    "github.com/ksred/llm/client"
    "github.com/ksred/llm/config"
)

func main() {
    // Configure the client
    cfg := &config.Config{
        Provider: "openai",
        Model:    "gpt-3.5-turbo",
        APIKey:   "your-api-key",
    }

    // Create a new client
    c, err := client.NewClient(cfg)
    if err != nil {
        panic(err)
    }

    // Send a chat request
    resp, err := c.Chat(context.Background(), &types.ChatRequest{
        Messages: []types.Message{
            {
                Role:    types.RoleUser,
                Content: "Hello, how are you?",
            },
        },
    })

    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Message.Content)
}
```

## Advanced Usage 🔧

### Streaming Responses
```go
streamChan, err := client.StreamChat(ctx, req)
if err != nil {
    return err
}

for resp := range streamChan {
    if resp.Error != nil {
        return resp.Error
    }
    fmt.Print(resp.Message.Content)
}
```

### Cost Tracking
```go
tracker := cost.NewCostTracker()
usage := types.Usage{
    PromptTokens:     100,
    CompletionTokens: 50,
}

// Track usage and cost
err := tracker.TrackUsage("openai", "gpt-3.5-turbo", usage)

// Get total cost
cost, err := tracker.GetCost("openai", "gpt-3.5-turbo")
fmt.Printf("Total cost: $%.4f\n", cost)
```

### Connection Pooling
```go
cfg := &config.Config{
    // ... other config
    PoolConfig: &resource.PoolConfig{
        MaxSize:       10,
        IdleTimeout:   5 * time.Minute,
        CleanupPeriod: time.Minute,
    },
}
```

## Examples 📚

The repository includes two example applications:

1. [Simple Example](examples/simple/README.md)
   - Basic chat functionality
   - Provider configuration
   - Error handling

2. [Advanced Example](examples/advanced/README.md)
   - Interactive CLI
   - Multiple providers
   - Streaming responses
   - Conversation management
   - Performance metrics
   - Cost tracking
   - Command system

## Architecture 🏗️

### Package Structure
- `client/` - Core client implementation
- `config/` - Configuration types and validation
- `models/` - Provider-specific implementations
- `pkg/` - Shared utilities and types
  - `cost/` - Cost tracking and budget management
  - `resource/` - Resource management (pools, retries)
  - `types/` - Common type definitions

### Key Components
1. **Client Interface**
   - Unified API for all providers
   - Stream and non-stream support
   - Context-aware operations

2. **Resource Management**
   - Connection pooling
   - Rate limiting
   - Automatic retries

3. **Cost Management**
   - Token counting
   - Usage tracking
   - Budget enforcement

4. **Error Handling**
   - Provider-specific error mapping
   - Retry strategies
   - Graceful degradation

## Contributing 🤝

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

### Development Prerequisites
- Go 1.20 or higher
- Make (for running development commands)
- API keys for testing

### Running Tests
```bash
make test
make integration  # Requires API keys in .env
```

## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments 🙏

- OpenAI for their GPT models and API
- Anthropic for their Claude models and API
- The Go community for excellent tooling and support
