# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG ARCH="amd64"
ARG OS="linux"
RUN CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -a -installsuffix cgo -o macropower_analytics_panel_server .

# Runtime stage
FROM quay.io/prometheus/busybox:latest

# Copy the binary from builder stage
COPY --from=builder /app/macropower_analytics_panel_server /bin/macropower_analytics_panel_server

USER nobody
ENTRYPOINT ["/bin/macropower_analytics_panel_server"]
EXPOSE 8080
