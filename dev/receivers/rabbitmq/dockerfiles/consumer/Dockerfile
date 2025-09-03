FROM golang:1.24.5-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd cmd
RUN go build ./cmd/consumer

# NOTE: Used with additional_contexts in Docker Compose.
FROM scratch AS ca-certificates

FROM scratch
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/consumer /
CMD ["./consumer"]
