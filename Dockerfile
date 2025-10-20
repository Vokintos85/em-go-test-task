# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations
EXPOSE 8080
ENV HTTP_PORT=8080
CMD ["./server"]
