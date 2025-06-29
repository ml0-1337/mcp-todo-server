# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'main.Version=$(git describe --tags --always --dirty)' -X 'main.BuildDate=$(date -u '+%Y-%m-%d %H:%M:%S')'" \
    -o mcp-todo-server .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S mcp && \
    adduser -u 1000 -S mcp -G mcp

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mcp-todo-server .

# Change ownership
RUN chown -R mcp:mcp /app

# Switch to non-root user
USER mcp

# Expose HTTP port
EXPOSE 8080

# Set default environment variables
ENV MCP_TRANSPORT=http
ENV MCP_PORT=8080
ENV MCP_HOST=0.0.0.0

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${MCP_PORT}/health || exit 1

# Run the application
ENTRYPOINT ["./mcp-todo-server"]
CMD ["-transport", "${MCP_TRANSPORT}", "-host", "${MCP_HOST}", "-port", "${MCP_PORT}"]