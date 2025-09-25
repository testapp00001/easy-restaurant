# --- Build Stage ---
# Use the official Golang image as a builder
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app, creating a static binary. CGO_ENABLED=0 is important for a lightweight final image.
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# --- Final Stage ---
# Use a minimal, non-root alpine image for the final container
FROM alpine:latest

# Set the working directory
WORKDIR /

# Copy the built binary from the builder stage
COPY --from=builder /server /server

# Copy the .env file. In a real production scenario, you would manage these as secrets.
COPY .env .

# Expose the port the app runs on
EXPOSE 3000

# Command to run the executable
CMD ["/server"]