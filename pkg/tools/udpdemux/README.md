# udpdemux

`udpdemux` listens on a specified UDP port and forwards all the packets it
receives onto a specified localhost ports.

## Usage

```
Listen for UDP traffic and forward to any number of localhost ports

./main [flags] [UDP localhost ports for forwarding...]

Flags:
  -listen-addr string
        UDP address to listen on (default "0.0.0.0:2004")

Example:
        ./main --listen-addr 0.0.0.0:8080 8081 8082
```
