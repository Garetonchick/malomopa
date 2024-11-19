FROM golang:1.23.3-alpine3.20
WORKDIR /

COPY . .
COPY ./internal/sources/sources-service.json ./config.json

RUN go mod download

CMD ["go", "run", "./internal/sources/", "-config", "./config.json"]
