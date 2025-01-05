# Stage 1: Build the Go application
FROM golang:1.23.4-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o goblast cmd/main.go

# Stage 2: Run the Go application
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl bash

WORKDIR /root/app

COPY --from=builder /app/goblast .
COPY --from=builder /app/configs ./configs
COPY build/go/wait-for-it.sh /wait-for-it.sh

EXPOSE 8080

CMD ["/wait-for-it.sh", "db:5432", "--", "/wait-for-it.sh", "nats:4222", "--", "./goblast"]
