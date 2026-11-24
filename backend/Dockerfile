# Стадия сборки
FROM golang:1.24.1-alpine3.21 AS builder

    WORKDIR /app

    COPY go.mod go.sum ./
    RUN go mod download

    COPY . .
    RUN CGO_ENABLED=0 go build -o main .

# Стадия выполнения
FROM alpine:3.21.3

    WORKDIR /app

    COPY --from=builder /app/main .

    EXPOSE 8080

    CMD ["./main"]

    HEALTHCHECK --interval=30s --timeout=1s --start-period=3s --retries=3 CMD /usr/bin/wget -q  http://localhost:8080/health/ready -O - | /bin/grep 'Ready check passed' || exit 1