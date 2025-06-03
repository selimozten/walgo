# Build stage
FROM golang:1.22-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o walgo main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S walgo && \
    adduser -u 1001 -S walgo -G walgo

# Set working directory
WORKDIR /home/walgo

# Copy binary from builder stage
COPY --from=builder /app/walgo /usr/local/bin/walgo

# Make sure binary is executable
RUN chmod +x /usr/local/bin/walgo

# Switch to non-root user
USER walgo

# Set entrypoint
ENTRYPOINT ["walgo"]

# Default command shows help
CMD ["--help"] 