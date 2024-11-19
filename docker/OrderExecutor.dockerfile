FROM golang:1.23.3-alpine3.20
WORKDIR /

# Копируем все файлы модуля в контейнер
# COPY go.mod go.sum ./

# COPY ./cmd/order-executor ./cmd/order-executor
# COPY ./internal ./internal

COPY . .
COPY ./example-config/order-executor.json ./config.json

RUN go mod download

CMD ["go", "run", "./cmd/order-executor/", "-config", "./config.json"]