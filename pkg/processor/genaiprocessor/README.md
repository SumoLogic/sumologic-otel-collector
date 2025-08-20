# GenAI Processor

The GenAI processor integrates with LiteLLM gateway to process OpenTelemetry log data using Large Language Models (LLMs). It sends log data to LLM models via API calls and enriches the logs with AI-generated insights.

## Configuration

The processor can be configured with the following parameters:

```yaml
processors:
  genai:
    # LiteLLM endpoint URL (required)
    endpoint: "http://localhost:4000"
    
    # API key for authentication (optional)
    api_key: "your-api-key"
    
    # Model to use for processing (default: "gpt-3.5-turbo")
    model: "gpt-3.5-turbo"
    
    # System prompt template (default: "You are a helpful assistant that analyzes log data.")
    system_prompt: "You are a log analysis expert. Provide concise insights about the following log data."
    
    # User prompt template with placeholders (default: "Analyze this log entry and provide insights: {{.body}}")
    user_prompt: "Analyze this log: {{.body}}. Severity: {{.severity}}. Timestamp: {{.timestamp}}"
    
    # Maximum tokens for the response (default: 150)
    max_tokens: 150
    
    # Temperature for response generation (default: 0.3)
    temperature: 0.3
    
    # Timeout for API requests (default: 30s)
    timeout: 30s
    
    # Field to store the AI response (default: "ai_analysis")
    response_field: "ai_analysis"
    
    # Only process logs matching this regex (optional)
    filter_regex: "ERROR|WARN"
    
    # Fields to extract from log records for processing (default: ["body"])
    extract_fields: ["body", "message", "error"]
```

## Features

- **Template-based prompts**: Use Go text templates to customize system and user prompts with log data
- **Selective processing**: Filter logs using regex patterns to process only relevant entries
- **Field extraction**: Extract specific fields from log attributes for use in prompts
- **Error handling**: Graceful error handling that doesn't stop the pipeline
- **Configurable models**: Support for any model available through LiteLLM
- **Response enrichment**: Add AI-generated insights as new attributes to log records

## Template Variables

The following variables are available in prompt templates:

- `{{.body}}` - Log message body
- `{{.timestamp}}` - Log timestamp
- `{{.severity}}` - Log severity level
- Any field specified in `extract_fields` configuration

## LiteLLM Integration

This processor expects a LiteLLM gateway running at the configured endpoint. LiteLLM should be configured with the appropriate model providers and API keys.

Example LiteLLM configuration:
```yaml
model_list:
  - model_name: gpt-3.5-turbo
    litellm_params:
      model: openai/gpt-3.5-turbo
      api_key: your-openai-api-key
```

## Example Configuration

```yaml
receivers:
  filelog:
    include: ["/var/log/*.log"]

processors:
  genai:
    endpoint: "http://litellm:4000"
    model: "gpt-3.5-turbo"
    system_prompt: "You are a security analyst. Analyze log entries for potential security issues."
    user_prompt: "Log: {{.body}}. Time: {{.timestamp}}. Is this suspicious?"
    max_tokens: 100
    temperature: 0.2
    response_field: "security_analysis"
    filter_regex: "(ERROR|CRITICAL|authentication|login|failed)"

exporters:
  logging:
    loglevel: info

service:
  pipelines:
    logs:
      receivers: [filelog]
      processors: [genai]
      exporters: [logging]
```

## Performance Considerations

- The processor makes HTTP requests to the LiteLLM API for each matching log record
- Use `filter_regex` to limit processing to relevant logs only
- Consider the `timeout` setting to prevent blocking the pipeline
- Monitor API rate limits and costs when using external LLM providers
- The processor processes logs sequentially to avoid overwhelming the API

## Security

- API keys should be stored securely (e.g., environment variables, secrets management)
- Be mindful of sensitive data in logs when sending to external LLM providers
- Consider using local models or private deployments for sensitive data
