# Go LLM Package

[![Go Reference](https://pkg.go.dev/badge/github.com/ksred/llm.svg)](https://pkg.go.dev/github.com/ksred/llm)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksred/llm)](https://goreportcard.com/report/github.com/ksred/llm)

A production-ready Go package for interacting with Large Language Models (LLMs), supporting multiple providers, streaming, cost management, and robust error handling.

## Features

- [ ] Multiple LLM Provider Support
  - [ ] OpenAI
  - [ ] Anthropic
  - [ ] Mock Provider for Testing
- [ ] Streaming Support
- [ ] Cost Management
- [ ] Rate Limiting
- [ ] Robust Error Handling
- [ ] Comprehensive Testing

## Installation

```bash
go get github.com/ksred/llm
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/ksred/llm"
)

func main() {
    // Example code will be added once the client implementation is complete
}
```

## Project Structure

The project follows standard Go project layout:

- `cmd/`: Command line tools
- `internal/`: Private application and library code
- `pkg/`: Library code that's ok to use by external applications
- `models/`: LLM provider implementations
- `examples/`: Examples for using the package
- `config/`: Configuration handling

## Development Status

This project is under active development. The checklist below tracks our progress:

- [x] Project Structure
- [ ] Core Interfaces
- [ ] Provider Implementations
- [ ] Testing Infrastructure
- [ ] Documentation
- [ ] CI/CD Pipeline

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
