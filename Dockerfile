# Build stage
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение из правильной директории
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/server .

EXPOSE 3000

CMD ["./server"]