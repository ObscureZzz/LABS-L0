# Используем Go 1.25.3 Alpine
FROM golang:1.25.3-alpine

# Рабочая директория
WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git bash

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник
RUN go build -o service .

# Экспортируем порт
EXPOSE 3000

# Запуск сервиса
CMD ["./service"]
