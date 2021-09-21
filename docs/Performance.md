# Performance

## Logs

### Benchmark setup

The following benchmark has been compiled on an Amazon `m4.large`
instance (which has 2 CPU cores and 8 GB of memory available).

It can be used when estimating the required CPU resources for logs collection
using [`filelogreceiver`][filelogreceiver].

[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver

#### CPU usage guidelines

#### Benchmark - CPU usage for particular average message size and EPS

Measured CPU usage for particular Events Per Second (EPS) average message size.

|          | 100B  | 512B  |  1KB   |  5KB   |  10KB  |
|:--------:|:-----:|:-----:|:------:|:------:|:------:|
| **EPS**  |   -   |   -   |   -    |   -    |   -    |
| **100**  | 1.14% |  1%   | 1.01%  |  1.4%  | 3.78%  |
| **200**  | 1.29% | 1.4%  | 1.41%  | 2.57%  | 5.36%  |
| **500**  | 2.75% | 2.71% | 2.95%  |  5.7%  | 10.68% |
| **1000** | 4.74% | 5.07% | 5.32%  | 11.3%  | 20.12% |
| **1500** | 7.08% | 7.29% | 7.99%  | 16.93% | 27.96% |
| **2000** | 9.64% | 9.56% | 10.39% | 22.51% | 36.59% |

#### Benchmark - EPS for average message size and CPU usage

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

#### Memory usage guidelines

#### Benchmark - memory usage for particular average message size and EPS

Measured memory usage (in MB) for particular Events Per Second (EPS) average message size.

|          |  100B  |  512B  |  1KB   |  5KB   |  10KB  |
|:--------:|:------:|:------:|:------:|:------:|:------:|
| **EPS**  |   -    |   -    |   -    |   -    |   -    |
| **100**  | 113.14 | 116.16 | 117.1  | 116.99 | 112.59 |
| **200**  | 115.16 | 118.55 | 116.8  | 119.67 | 127.02 |
| **500**  | 118.24 | 121.79 | 122.78 | 127.87 | 142.73 |
| **1000** | 121.6  | 126.75 | 127.94 | 140.11 | 106.82 |
| **1500** | 128.54 | 131.9  | 137.69 | 95.21  | 113.89 |
| **2000** | 130.62 | 125.27 | 144.59 | 98.62  | 134.61 |
