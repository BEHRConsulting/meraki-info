# Retry Logic Implementation

## Overview

The Meraki Info CLI application now includes robust retry logic for API calls to handle transient failures, rate limiting, and server errors. This implementation helps ensure reliable data collection even when the Meraki API experiences temporary issues.

## Features

### Automatic Retry on Transient Errors
- **Network errors**: Connection timeouts, DNS failures, etc.
- **Rate limiting (429)**: Automatic retry with exponential backoff
- **Server errors (5xx)**: Retry on 500, 502, 503, 504 status codes
- **No retry on client errors**: 400, 401, 403, 404 are not retried

### Exponential Backoff
- **Initial interval**: 1 second (configurable)
- **Maximum interval**: 30 seconds (configurable)
- **Multiplier**: 2.0 (configurable)
- **Maximum retries**: 3 (configurable)

### Configurable Retry Behavior
Users can customize retry behavior through the client configuration.

## Default Configuration

```go
RetryConfig{
    MaxRetries:      3,
    InitialInterval: 1 * time.Second,
    MaxInterval:     30 * time.Second,
    Multiplier:      2.0,
}
```

## Usage Examples

### Basic Usage (Default Configuration)
```go
client, err := meraki.NewClient("your-api-key")
if err != nil {
    log.Fatal(err)
}

// API calls will automatically retry on transient failures
organizations, err := client.GetOrganizations()
```

### Custom Retry Configuration
```go
client, err := meraki.NewClient("your-api-key")
if err != nil {
    log.Fatal(err)
}

// Configure custom retry behavior
client.SetRetryConfig(meraki.RetryConfig{
    MaxRetries:      5,                    // More retries
    InitialInterval: 500 * time.Millisecond, // Shorter initial wait
    MaxInterval:     60 * time.Second,     // Longer maximum wait
    Multiplier:      2.5,                  // More aggressive backoff
})

// API calls will use the custom retry configuration
organizations, err := client.GetOrganizations()
```

### Checking Current Configuration
```go
client, err := meraki.NewClient("your-api-key")
if err != nil {
    log.Fatal(err)
}

config := client.GetRetryConfig()
fmt.Printf("Max retries: %d\n", config.MaxRetries)
fmt.Printf("Initial interval: %v\n", config.InitialInterval)
```

## Retry Behavior Details

### Retryable Errors
- **Network errors**: Any error during the HTTP request
- **HTTP 429**: Too Many Requests (rate limiting)
- **HTTP 500**: Internal Server Error
- **HTTP 502**: Bad Gateway
- **HTTP 503**: Service Unavailable
- **HTTP 504**: Gateway Timeout

### Non-Retryable Errors
- **HTTP 400**: Bad Request
- **HTTP 401**: Unauthorized
- **HTTP 403**: Forbidden
- **HTTP 404**: Not Found
- **HTTP 2xx**: Success responses

### Backoff Calculation
The backoff duration is calculated using exponential backoff:
```
backoff = initial_interval * (multiplier ^ (attempt - 1))
```

The calculated backoff is capped at the maximum interval to prevent excessively long waits.

## Logging

The retry logic includes detailed logging:
- **Debug logs**: Every API request attempt
- **Info logs**: Retry attempts with error details and backoff duration
- **Error logs**: Final failure after all retries exhausted

Example log output:
```
2025/07/17 06:28:19 INFO Request failed with retryable status, retrying status=429 attempt=1 backoff=1s
2025/07/17 06:28:20 INFO Request failed with retryable status, retrying status=429 attempt=2 backoff=2s
2025/07/17 06:28:22 DEBUG API request successful method=GET url=https://api.meraki.com/api/v1/organizations attempt=3
```

## Performance Considerations

### Rate Limiting
- The Meraki API has rate limits that vary by endpoint
- The retry logic automatically handles 429 responses
- Exponential backoff reduces the likelihood of repeated rate limit hits

### Timeout Handling
- The HTTP client has a 30-second timeout
- Total retry time can be calculated as: `sum of all backoff periods + (max_retries * request_timeout)`
- With default settings, maximum total time is approximately 1 + 2 + 4 + (3 * 30) = 97 seconds

### Memory Usage
- Retry logic uses minimal additional memory
- No request body buffering is required since all requests are GET requests

## Error Handling

### Retry Exhaustion
When all retries are exhausted, the client returns a detailed error message:
```
API request failed with status 500 after 4 attempts
```

### Network Errors
Network errors are wrapped with context about the number of attempts:
```
failed to make request after 4 attempts: connection timeout
```

## Testing

The retry logic includes comprehensive tests covering:
- Successful requests on first attempt
- Retry on rate limiting (429)
- Retry on server errors (5xx)
- No retry on client errors (4xx)
- Retry exhaustion scenarios
- Backoff calculation verification
- Configuration management

## Best Practices

1. **Use default configuration** for most use cases
2. **Monitor logs** to understand retry patterns
3. **Adjust retry configuration** based on your specific needs:
   - Increase max retries for critical operations
   - Decrease initial interval for time-sensitive operations
   - Increase max interval for long-running batch operations
4. **Handle final errors** appropriately in your application
5. **Consider circuit breaker patterns** for additional resilience in high-volume scenarios

## Migration Notes

The retry logic is implemented at the HTTP client level and is transparent to existing code. No changes are required to existing API calls - they will automatically benefit from the retry logic.

## Performance Impact

- **Minimal overhead**: Retry logic only activates on failures
- **Improved reliability**: Reduces transient failure rates
- **Predictable latency**: Exponential backoff provides bounded retry times
- **Rate limit friendly**: Automatic handling of 429 responses
