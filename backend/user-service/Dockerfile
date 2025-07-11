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
# Install swag CLI tool
RUN go install github.com/swaggo/swag/cmd/swag@latest
# Run swag init to generate docs (replace module path with your actual module path)
# e.g., if your module is github.com/your_username/backend/user-service
RUN swag init --parseDependency --parseInternal -g ./main.go --output docs

# Build the Go application
# -o /app/user-service specifies the output file name
# -ldflags "-s -w" reduces the binary size and removes debug information for smaller image
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./main.go

# Use a minimal image for the final stage to reduce image size
FROM alpine:latest

# Set the working directory to root
WORKDIR /root/

# Copy the compiled application from the builder stage
COPY --from=builder /app/user-service .
# Copy Swagger documentation (needed if you want Swagger UI to work in the container)
COPY --from=builder /app/docs ./docs/

# Expose the port the application runs on
EXPOSE 8080

# Run the executable
CMD ["./user-service"]