# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed for private repos and HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

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
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o vless-generator .

# Final stage
FROM scratch

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy our binary
COPY --from=builder /app/vless-generator /vless-generator

# Copy templates
COPY --from=builder /app/templates /templates

# Create a non-root user
# Note: In scratch image, we can't use adduser, so we'll run as root
# For better security, consider using alpine:latest as base instead

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/vless-generator", "-help"] || exit 1

# Set environment variables
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json
ENV PORT=8080

# Run the binary
ENTRYPOINT ["/vless-generator"]
CMD ["-log-level", "info", "-log-format", "json", "-port", "8080"]
