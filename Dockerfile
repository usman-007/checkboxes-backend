FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/checkbox-backend .

# Use a minimal alpine image for the final container
FROM alpine:3.18

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/checkbox-backend /app/checkbox-backend

# Copy any config files if needed
# COPY config/ /app/config/

# Expose the port your application runs on
EXPOSE 8080

# Command to run the executable
CMD ["/app/checkbox-backend"]
