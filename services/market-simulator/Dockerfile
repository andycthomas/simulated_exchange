# Build stage
FROM golang:1.23 AS builder

# Update CA certificates and disable SSL verification
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Set working directory
WORKDIR /app

# Copy source code first (includes go.mod, go.sum)
COPY . .

# Set build environment
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOPROXY=direct
ENV GOSUMDB=off
ENV GIT_SSL_NO_VERIFY=1
ENV GOINSECURE='*'

# Skip module download - build directly with modules from source

# Build the market simulator service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o market-simulator ./services/market-simulator/cmd

# Build health check binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o healthcheck ./cmd/healthcheck

# Final stage - Distroless base
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/market-simulator .
COPY --from=builder /app/healthcheck .

# Copy SSL certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 8081

# Health check (Note: distroless doesn't have wget, health check handled by orchestrator)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./market-simulator"]