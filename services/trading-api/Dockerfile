# Build stage
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Set Go proxy to direct and disable SSL verification to bypass host SSL issues
ENV GOPROXY=direct
ENV GOSUMDB=off
ENV GIT_SSL_NO_VERIFY=1
ENV GOINSECURE='*'

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the trading API service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o trading-api ./services/trading-api/cmd

# Build health check binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o healthcheck ./cmd/healthcheck

# Final stage - Distroless base
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/trading-api .
COPY --from=builder /app/healthcheck .

# Copy SSL certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 8080

# Health check (Note: distroless doesn't have wget, health check handled by orchestrator)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./trading-api"]