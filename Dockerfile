# Stage 1: Build Go binary using lightweight Alpine
FROM golang:1.24-alpine AS builder

# Set workdir
WORKDIR /app

# Copy go.mod/go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Go binary
RUN GOOS=linux go build -o koito-multi-proxy .

# Stage 2: Runtime on Ubuntu 24.04
FROM ubuntu:24.04

# Avoid interactive prompts during apt installs
ENV DEBIAN_FRONTEND=noninteractive

# Copy Go binary from builder stage
COPY --from=builder /app/koito-multi-proxy /usr/local/bin/koito-multi-proxy

# Entrypoint
ENTRYPOINT ["koito-multi-proxy"]
