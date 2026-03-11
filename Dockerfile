# Этап сборки бинарника
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем модули и зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статически связанный бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server .


# Этап рантайма
FROM alpine:latest

WORKDIR /app

# Переменные окружения по умолчанию
ENV TODO_PORT=7540 \
    TODO_DBFILE=/app/scheduler.db \
    TODO_PASSWORD=""

# Копируем бинарник и фронтенд
COPY --from=builder /app/server /app/server
COPY web /app/web

# Порт, который слушает сервер
EXPOSE 7540

# Команда запуска приложения
CMD ["/app/server"]

