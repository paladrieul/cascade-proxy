# cascade-proxy

Lightweight HTTP proxy with automatic retry and circuit breaker patterns for microservices.

## Installation

```bash
go install github.com/cascade-proxy/cascade-proxy@latest
```

## Usage

Start the proxy with a target upstream service:

```bash
cascade-proxy --port 8080 --upstream http://your-service:3000
```

Or use it as a library in your Go project:

```go
package main

import (
    "github.com/cascade-proxy/cascade-proxy/pkg/proxy"
)

func main() {
    p := proxy.New(proxy.Config{
        Upstream:       "http://your-service:3000",
        Port:           8080,
        MaxRetries:     3,
        CircuitBreaker: true,
    })

    p.Start()
}
```

### Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | Port to listen on |
| `--upstream` | required | Target upstream URL |
| `--max-retries` | `3` | Max retry attempts on failure |
| `--timeout` | `30s` | Request timeout |
| `--cb-threshold` | `5` | Circuit breaker failure threshold |

## Features

- Automatic request retries with exponential backoff
- Circuit breaker to prevent cascading failures
- Configurable timeouts and retry policies
- Minimal overhead and low memory footprint

## License

MIT © cascade-proxy contributors