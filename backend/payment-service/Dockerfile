# Use the official Golang image as a base image for building
FROM golang:1.23.4-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Generate Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --parseDependency --parseInternal -g ./main.go --output docs

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/payment-service ./main.go

# Use a minimal image for the final stage to reduce image size
FROM alpine:latest

# Set the working directory to root
WORKDIR /root/

# Copy the compiled application from the builder stage
COPY --from=builder /app/payment-service .
COPY --from=builder /app/docs ./docs/

# Expose the port the application runs on
EXPOSE 8082

# Run the executable
CMD ["./payment-service"]