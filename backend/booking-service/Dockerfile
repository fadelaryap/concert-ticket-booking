FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN apk add --no-cache netcat-openbsd git ca-certificates openssl-dev

RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --parseDependency --parseInternal -g ./main.go --output docs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/booking-service ./main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/booking-service .
COPY --from=builder /app/docs ./docs/

COPY ./scripts/wait-for-it.sh /usr/local/bin/wait-for-it.sh
RUN chmod +x /usr/local/bin/wait-for-it.sh # Pastikan script bisa dieksekusi

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8081

CMD ["/usr/local/bin/wait-for-it.sh", "mysql", "3306", "--", "./booking-service"]
