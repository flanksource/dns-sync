FROM golang:1.24-alpine AS builder

ARG TARGETARCH

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -extldflags '-static'" \
    -a -installsuffix cgo \
    -o dns-sync cmd/dns-sync/main.go

# Final stage
FROM scratch AS production

ARG TARGETARCH

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/dns-sync /dns-sync

# Create directories for potential volume mounts
COPY --from=builder --chown=65534:65534 /tmp /tmp

# Use non-root user
USER 65534:65534

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/dns-sync", "--version"] || exit 1

EXPOSE 5353/udp
ENTRYPOINT ["/dns-sync"]
CMD ["-config", "/app/config.yaml"]

