# Build Stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o crochet-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o crochet-cli ./cmd/cli

# Run Stage
FROM alpine:latest

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binaries
COPY --from=builder /app/crochet-server .
COPY --from=builder /app/crochet-cli .

# Copy assets
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

# Create data directories and set permissions
RUN mkdir -p /app/data && \
    mkdir -p /app/static/uploads && \
    chown -R appuser:appgroup /app

# Expose port
EXPOSE 8585

# Set environment variables
ENV PORT=8585
ENV DB_PATH=/app/data/crochet.db

# Switch to non-root user
USER appuser

CMD ["./crochet-server"]