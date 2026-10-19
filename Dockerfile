# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Service name to build (passed via --build-arg)
ARG SERVICE_NAME

# Build the specific service
# CGO_ENABLED=0 creates a statically linked binary
RUN if [ -d "./services/${SERVICE_NAME}/cmd" ]; then \
      CGO_ENABLED=0 GOOS=linux go build -o /app/main ./services/${SERVICE_NAME}/cmd; \
    else \
      CGO_ENABLED=0 GOOS=linux go build -o /app/main ./services/${SERVICE_NAME}; \
    fi

# Production stage
FROM alpine:latest

# Add ca-certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
