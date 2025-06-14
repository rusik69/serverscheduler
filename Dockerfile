# Build frontend
FROM node:18-alpine as frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Build backend
FROM golang:1.21-alpine as backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:latest
WORKDIR /app

# Install SQLite
RUN apk add --no-cache sqlite

# Copy frontend build
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Copy backend binary
COPY --from=backend-builder /app/server .

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"] 