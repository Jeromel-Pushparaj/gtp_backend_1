# 🌟 GTP Backend API Gateway - Dockerfile
# Multi-stage build for optimal image size

# ═══════════════════════════════════════════════════════════
# 🔨 Stage 1: Build the application
# ═══════════════════════════════════════════════════════════
FROM golang:1.21-alpine AS builder

# 📦 Install build dependencies
RUN apk add --no-cache git make

# 📁 Set working directory
WORKDIR /app

# 📋 Copy go mod files
COPY go.mod go.sum ./

# 📦 Download dependencies
RUN go mod download

# 📁 Copy source code
COPY . .

# 🔨 Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway main.go

# ═══════════════════════════════════════════════════════════
# 🚀 Stage 2: Create minimal runtime image
# ═══════════════════════════════════════════════════════════
FROM alpine:latest

# 🔐 Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# 📁 Create app directory
WORKDIR /root/

# 📋 Copy binary from builder
COPY --from=builder /app/api-gateway .

# 📋 Copy .env.example (optional - for reference)
COPY --from=builder /app/.env.example .

# 🌐 Expose gateway port
EXPOSE 8000

# 🏷️ Add labels
LABEL maintainer="GTP Backend Team"
LABEL description="API Gateway for GTP Backend Microservices"
LABEL version="1.0.0"

# 🚀 Run the gateway
CMD ["./api-gateway"]

