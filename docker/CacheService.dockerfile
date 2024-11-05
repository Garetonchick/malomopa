FROM golang:1.23.2-alpine3.20
WORKDIR cache-service/
COPY cache-service/ .
CMD ["go", "run", "./cmd/cache-service"]
EXPOSE 4444
