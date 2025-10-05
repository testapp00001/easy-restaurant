# file: Dockerfile

# =================================================================
# 1. Development Stage (`development`)
# This stage includes the Go toolchain and air for live reloading.
# =================================================================
FROM golang:1.25-alpine AS development

# Install air for live reloading
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# This is the command that will be run by our docker-compose dev setup
CMD ["air", "-c", ".air.toml"]


# =================================================================
# 2. Builder Stage (`builder`)
# This stage is for compiling the production binary.
# =================================================================
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the static binary for production
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server


# =================================================================
# 3. Production Stage (`production`)
# This is the final, minimal image for deployment.
# =================================================================
FROM alpine:latest AS production

WORKDIR /

# Copy only the built binary from the builder stage
COPY --from=builder /server /server

# Expose the port the app runs on
EXPOSE 3000

# Command to run the executable
CMD ["/server"]