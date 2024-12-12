# Build stage
FROM golang:1.21-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /reconciliation ./cmd/main.go

# Final stage
FROM alpine:3.19

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /reconciliation .

# Create a directory for input files
RUN mkdir -p /app/data

# Set the entrypoint
ENTRYPOINT ["/app/reconciliation"] 