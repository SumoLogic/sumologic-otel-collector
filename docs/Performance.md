# Performance

## Logs

### CPU usage guidelines

#### Setup

The following benchmark has been compiled on an Amazon `m4.large`
instance (which has 2 CPU cores and 8 GB of memory available).

It can be used when estimating the required CPU resources for logs collection
using [`filelogreceiver`][filelogreceiver].

[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver

#### Benchmark

Events Per Second (EPS) achieved for a particular average message size and CPU usage.

|                       | 100B  | 512B  | 1KB  | 5KB  | 10KB |
|:---------------------:|:-----:|:-----:|:----:|:----:|:----:|
| **Average CPU usage** |   -   |   -   |  -   |  -   |  -   |
|        **5%**         | 2000  | 1100  | 1000 | 150  | 200  |
|        **10%**        | 3500  | 2100  | 1500 | 450  | 300  |
|        **20%**        | 6500  | 4100  | 3000 | 1200 | 700  |
|        **50%**        | 14000 | 10100 | 8500 |  -*  |  -*  |
|        **90%**        |  -*   | 19100 |  -*  |  -*  |  -*  |

\* - cells without a resulting EPS come from the fact that the CPU utilization
didn't reach the designated CPU utilization during the benchmark run.

The above table can be interpreted in the following way:

For an average CPU usage of 5%

- 10 KB logs can be ingested at 200 logs/sec (2000 KB/sec).
- 1 KB logs can be ingested at 1000 logs/sec (1000 KB/sec).

This shows that the collector performs better when it is made to ingest bigger
log entries (which is expected due to less overhead coming from timestamp parsing etc.).
