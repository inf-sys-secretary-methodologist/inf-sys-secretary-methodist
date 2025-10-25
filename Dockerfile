# backend/Dockerfile

# Stage 1: Build
FROM golang:1.25.3-bookworm AS builder

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статический бинарник (без CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Runtime на Ubuntu
FROM ubuntu:22.04

# Обновляем систему и устанавливаем минимальные зависимости
RUN apt-get update && \
    apt-get install -y ca-certificates curl tzdata && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/server .

# Порты: HTTP и gRPC
EXPOSE 8080 

# Запуск
CMD ["./main"]