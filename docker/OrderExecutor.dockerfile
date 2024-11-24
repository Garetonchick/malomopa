FROM golang:1.23.3-alpine3.20
WORKDIR /

COPY . .
COPY ./example-config/order-executor.json ./config.json

RUN go mod download

CMD ["go", "run", "./cmd/order-executor/", "-config", "./config.json"]