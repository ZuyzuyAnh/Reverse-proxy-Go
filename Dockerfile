# Build stage
FROM golang:1.23.5 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reverse-proxy ./cmd/

# Runtime stage
FROM alpine:latest
WORKDIR /root
COPY --from=builder /app/reverse-proxy .
RUN chmod +x /root/reverse-proxy
CMD ["/root/reverse-proxy"]
