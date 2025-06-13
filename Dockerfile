# Build stage
FROM golang:1.22-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/server ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates and SQLite
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Create data directory for SQLite database
RUN mkdir -p /app/data

# Copy the binary from builder
COPY --from=builder /app/bin/server .

# Expose port 8080
EXPOSE 8080

# Set environment variable for database path
ENV DB_PATH=/app/data/server_scheduler.db

# Run the application
CMD ["./server"] 