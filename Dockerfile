# Build Stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o vyakhya main.go

# Production Stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/vyakhya /app/vyakhya

# Copy WordNet dataset (needed at runtime)
COPY --from=builder /app/english-wordnet /app/english-wordnet

# Expose server port
EXPOSE 8080

# Run the server
ENTRYPOINT ["/app/vyakhya", "-port", "8080", "-dataset", "/app/english-wordnet"]
