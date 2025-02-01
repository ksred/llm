# Advanced LLM Client Example

This example demonstrates advanced features of the LLM client library, including:

- Interactive chat with multiple providers (OpenAI and Anthropic)
- Real-time streaming responses
- Conversation history management
- Performance metrics tracking
- Cost tracking and budgeting
- Command handling
- Graceful shutdown

## Features

### Multiple Providers
- Supports both OpenAI (GPT-3.5) and Anthropic (Claude) models
- Runs queries in parallel across all available providers
- Color-coded responses for easy differentiation

### Conversation Management
- Maintains conversation history for each provider
- Supports clearing history with `/clear` command
- Properly handles context and message threading

### Performance Metrics
- Tracks request counts, success/failure rates
- Measures response times and latency
- Monitors token usage (input/output)
- Available via `/metrics` command

### Cost Tracking
- Real-time cost calculation based on token usage
- Separate tracking for prompt and completion tokens
- Support for provider-specific pricing (GPT-3.5, Claude-2.1)
- View total costs via `/metrics` command

### Commands
- `/help` - Show available commands
- `/metrics` - Display performance metrics and costs
- `/clear` - Clear conversation history
- `/quit` - Exit gracefully

## Usage

1. Set up environment variables in `.env`:
   ```
   OPENAI_API_KEY=your_openai_key
   ANTHROPIC_API_KEY=your_anthropic_key
   ```

2. Run the example:
   ```
   go run examples/advanced/main.go
   ```

3. Start chatting! The message will be sent to all available providers.

4. Use commands like `/metrics` to view performance stats and costs.

## Error Handling

- Graceful handling of API errors
- Proper cleanup on shutdown
- Timeout handling for requests
- Budget tracking and warnings

## Implementation Details

- Uses provider-specific token counting
- Real-time cost calculation using current API pricing
- Concurrent request handling with proper context management
- Signal handling for clean shutdowns
