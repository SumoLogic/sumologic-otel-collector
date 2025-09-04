FROM golang:1.24.5-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
COPY cmd cmd
RUN go build ./cmd/publisher

# NOTE: Used with additional_contexts in Docker Compose.
FROM scratch AS ca-certificates

FROM scratch
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/publisher /
CMD ["./publisher"]
