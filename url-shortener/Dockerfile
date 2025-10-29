# Стадия сборки: компилируем Go приложение
FROM golang:1.25.3-alpine AS builder

# Устанавливаем зависимости для компиляции (если нужны C-библиотеки)
RUN apk add --no-cache git ca-certificates

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей сначала (для кэширования слоев)
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Компилируем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener ./cmd/server

# Стадия запуска: минимальный образ
FROM alpine:latest

# Устанавливаем зависимости для runtime
RUN apk add --no-cache ca-certificates

# Создаем пользователя для безопасности
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Переключаемся на непривилегированного пользователя
USER appuser

# Копируем скомпилированное приложение из стадии builder
COPY --from=builder --chown=appuser:appgroup /app/url-shortener /app/

# Указываем рабочую директорию
WORKDIR /app

# Объявляем порт, который использует приложение
EXPOSE 8080

# Команда запуска приложения
CMD ["./url-shortener"]