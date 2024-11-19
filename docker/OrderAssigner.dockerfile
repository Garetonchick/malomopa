FROM golang:1.23.3-alpine3.20
WORKDIR /

COPY . .
COPY ./example-config/order-assigner.json ./config.json

RUN go mod download

CMD ["go", "run", "./cmd/order-assigner/", "-config", "./config.json"]