FROM golang:1.23.3-alpine3.20
WORKDIR /

COPY . .
COPY ./example-config/sources-service.json ./config.json

RUN go mod download

CMD ["go", "run", "./cmd/sources/", "-config", "./config.json"]
