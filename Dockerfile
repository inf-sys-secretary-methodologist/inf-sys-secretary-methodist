# backend/Dockerfile

# Stage 1: Build
FROM golang:1.26.0-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Копируем зависимости и скачиваем их (кэшируется отдельным слоем)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Копируем только необходимый код
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/swagger/ ./docs/swagger/

# Собираем статический бинарник с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Stage 2: Минимальный runtime на scratch
FROM scratch

# Копируем необходимые системные файлы
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копируем бинарник
COPY --from=builder /app/server /server

# Порт HTTP
EXPOSE 8080

# Запуск
ENTRYPOINT ["/server"]