# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies for CGO and SQLite compilation
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with musl-compatible flags for go-sqlite3
# -D_LARGEFILE64_SOURCE enables 64-bit file I/O compatibility in musl
RUN CGO_ENABLED=1 GOOS=linux \
    CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
    go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Copy binary
COPY --from=builder /app/server .

# Create data directory
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./server"]
