# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make protobuf-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taskqueue .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/taskqueue .

# Copy configuration and web assets
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/web ./web

# Expose ports
EXPOSE 8080 9090 2112

# Default command - run in both mode
CMD ["./taskqueue", "both"]
