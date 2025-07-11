
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN swag init --parseDependency --parseInternal -g ./main.go --output docs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/user-service .
COPY --from=builder /app/docs ./docs/

EXPOSE 8080

CMD ["./user-service"]
