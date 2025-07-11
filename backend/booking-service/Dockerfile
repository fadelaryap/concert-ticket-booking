# Use the official Golang image as a base image for building
FROM golang:1.23.4-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

RUN apk add --no-cache netcat-openbsd git ca-certificates openssl-dev

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Generate Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --parseDependency --parseInternal -g ./main.go --output docs

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/booking-service ./main.go

# Use a minimal image for the final stage to reduce image size
FROM alpine:latest

# Set the working directory to root
WORKDIR /root/

# Copy the compiled application from the builder stage
COPY --from=builder /app/booking-service .
COPY --from=builder /app/docs ./docs/

COPY ./scripts/wait-for-it.sh /usr/local/bin/wait-for-it.sh
RUN chmod +x /usr/local/bin/wait-for-it.sh # Pastikan script bisa dieksekusi

# Copy CA certificates for Go's HTTPS/TLS connections (e.g., to RabbitMQ if using amqps)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose the port the application runs on
EXPOSE 8081

# Run the executable
CMD ["/usr/local/bin/wait-for-it.sh", "mysql", "3306", "--", "./booking-service"]