FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies first (caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server/main.go

# Production stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy built binary from builder
COPY --from=builder /app/bin/server /app/server

# Expose API port
EXPOSE 8080

CMD ["/app/server"]
