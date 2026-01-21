# DNS Lookup Source

This package provides DNS lookup capabilities for the lookup processor, supporting both forward (hostname to IP) and reverse (IP to hostname) DNS resolution with optional caching.

## Features

- **Forward DNS Lookups**: Resolve hostnames to IP addresses (A/AAAA records)
- **Reverse DNS Lookups**: Resolve IP addresses to hostnames (PTR records)
- **IPv4 and IPv6 Support**: Handles both IP versions seamlessly
- **Custom DNS Servers**: Configure specific DNS resolvers
- **Caching**: Optional in-memory caching with configurable TTL
- **Multiple Results**: Return all DNS results or just the first match
- **Configurable Timeouts**: Control DNS query timeout duration

## Configuration

```yaml
lookup:
  sources:
    - type: dns
      mode: forward              # "forward" or "reverse"
      timeout: 5s                # DNS query timeout
      resolver: "8.8.8.8:53"     # Optional: custom DNS server
      multiple_results: false    # Return all results or first only
```

### Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | string | `forward` | DNS resolution mode: `forward` (hostname→IP) or `reverse` (IP→hostname) |
| `timeout` | duration | `5s` | Maximum time to wait for DNS resolution |
| `resolver` | string | system default | Custom DNS server in `host:port` format (e.g., `8.8.8.8:53`) |
| `multiple_results` | bool | `false` | If true, returns all results as comma-separated string; if false, returns first result only |

## Performance Benchmarks

The following benchmarks demonstrate DNS lookup performance with and without caching.

### Test Environment

- **CPU**: AMD EPYC 7763 64-Core Processor
- **Memory**: 16 GB RAM
- **OS**: Linux (Ubuntu 24.04.3 LTS)
- **Go Version**: go1.25.4 linux/amd64
- **Architecture**: amd64

### Benchmark Results

```
BenchmarkDNSLookup/PTR_no_cache-4              24789     244403 ns/op     3648 B/op     34 allocs/op
BenchmarkDNSLookup/PTR_with_cache-4         34027570        175.3 ns/op     128 B/op      1 allocs/op
BenchmarkDNSLookupParallel/with_cache-4     58234840         98.37 ns/op     128 B/op      1 allocs/op
BenchmarkCacheEffectiveness/uncached-4         25374     235425 ns/op     3648 B/op     34 allocs/op
BenchmarkCacheEffectiveness/cached-4        34393330        175.6 ns/op     128 B/op      1 allocs/op
```

### Performance Analysis

| Scenario | Latency | Throughput | Memory per Op | Allocations |
|----------|---------|------------|---------------|-------------|
| **Uncached DNS** | ~240 μs (0.24 ms) | ~4,100 ops/sec | 3,648 B | 34 |
| **Cached DNS** | ~175 ns (0.00017 ms) | ~5.7M ops/sec | 128 B | 1 |
| **Parallel Cached** | ~98 ns | ~10M ops/sec | 128 B | 1 |

### Recommendations

1. **Enable caching** for production workloads to achieve maximum performance
2. **Uncached lookups** at ~240 μs are suitable for <4,000 queries/sec
3. **Cached lookups** can handle millions of queries per second
4. Configure appropriate cache TTL based on DNS record stability

## Running Benchmarks

To run the benchmarks yourself:

```bash
# Run all benchmarks with 10 second duration
go test -bench=. -benchtime=10s

# Run with memory statistics
go test -bench=. -benchtime=10s -benchmem

# Run specific benchmark
go test -bench=BenchmarkDNSLookup -benchtime=10s

# Run with CPU profiling
go test -bench=. -benchtime=10s -cpuprofile=cpu.prof

# Note: Benchmarks are skipped in short mode
# To skip: go test -short
```

## Testing

Run the test suite:

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestIPv6Support

# Run tests with coverage
go test -cover

# Skip long-running benchmarks
go test -short