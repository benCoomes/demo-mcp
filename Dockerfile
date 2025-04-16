FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Use a small alpine image for the final container
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose the port the server runs on
EXPOSE 8080

# Use an entrypoint to allow for different args
ENTRYPOINT ["./server"]
# Default to HTTP mode (empty CMD) - for stdio use "stdio" arg
CMD []