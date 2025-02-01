# Simple LLM Client Example

This example demonstrates how to use the LLM client library with both OpenAI and Anthropic providers. It shows:
1. Regular chat completion
2. Streaming chat completion

## Prerequisites

1. Create a `.env` file in the project root with your API keys:
```
OPENAI_API_KEY=your_openai_key_here
ANTHROPIC_API_KEY=your_anthropic_key_here
```

2. Make sure you have the required dependencies:
```bash
go mod tidy
```

## Running the Example

From the project root:
```bash
go run examples/simple/main.go
```

This will:
1. Test regular chat with both providers
2. Test streaming chat with both providers

Each test will show the provider name and response.
