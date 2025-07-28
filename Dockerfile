FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git build-base

# Устанавливаем migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /url-shortener

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener/url-shortener ./cmd/url-shortener/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /url-shortener

# Устанавливаем инструменты для работы с БД
RUN apk add --no-cache postgresql-client bash

# Копируем бинарники, миграции, конфиг и скрипт запуска
COPY --from=builder /url-shortener/url-shortener .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY ./migrations ./migrations
COPY ./config ./config
COPY ./entrypoint.sh .

# Делаем скрипт исполняемым внутри контейнера
RUN chmod +x ./entrypoint.sh

EXPOSE 8082

# Устанавливаем команду по умолчанию
CMD ["./entrypoint.sh"]