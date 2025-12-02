# Multi-stage build для pg-sec-lab
FROM golang:1.25-alpine AS builder

# Установка необходимых инструментов
RUN apk add --no-cache git

WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pg-sec-lab .

# Финальный образ
FROM alpine:latest

# Установка PostgreSQL клиента для дополнительных тестов
RUN apk add --no-cache postgresql-client ca-certificates

WORKDIR /app

# Копируем скомпилированное приложение
COPY --from=builder /app/pg-sec-lab .
COPY --from=builder /app/policy.yaml .

# Создаем директорию для вывода
RUN mkdir -p /app/output

# Копируем тестовые скрипты
COPY docker-test.sh /app/
RUN chmod +x /app/docker-test.sh

# По умолчанию показываем справку
CMD ["/app/pg-sec-lab", "--help"]
