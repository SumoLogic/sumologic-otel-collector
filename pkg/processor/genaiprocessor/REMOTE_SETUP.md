# GenAI Processor with Remote LiteLLM Gateway

This guide shows how to use the GenAI processor with a remote LiteLLM gateway instance, eliminating the need for local LiteLLM setup.

## Benefits of Remote Setup

**No local installation** - No need to install and manage LiteLLM locally  
**Shared resources** - Multiple collectors can use the same gateway  
**Centralized management** - Models and configurations managed in one place  
**Better scaling** - Remote gateway can be optimized for performance  
**Simplified deployment** - Collectors only need the GenAI processor configuration  

## Quick Start

### 1. Configure the endpoint

Edit `remote_config.yaml` and update the endpoint:

```yaml
processors:
  genai:
    endpoint: "https://your-litellm-gateway.example.com"
    # api_key: "${LITELLM_API_KEY}"  # Uncomment if authentication required
```

### 2. Set up authentication (if required)

```bash
export LITELLM_API_KEY="your-api-key"
```

Then uncomment the `api_key` line in the configuration.

### 3. Test the connection

```bash
./test_remote.sh
```

### 4. Run the collector

```bash
otelcolbuilder/cmd/otelcol-sumo-* --config=remote_config.yaml
```

## Configuration Options

### Basic Configuration (No Auth)

```yaml
processors:
  genai:
    endpoint: "https://your-gateway.com"
    model: "gpt-3.5-turbo"
    # ... other settings
```

### With API Key Authentication

```yaml
processors:
  genai:
    endpoint: "https://your-gateway.com"
    api_key: "${LITELLM_API_KEY}"
    model: "gpt-3.5-turbo"
    # ... other settings
```
