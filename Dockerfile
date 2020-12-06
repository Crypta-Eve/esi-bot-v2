FROM golang:1.15.2 as builder
WORKDIR /app
COPY . .
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bot ./cmd/eb2

FROM alpine:latest AS release
WORKDIR /app/logs
WORKDIR /app

RUN apk --update --no-cache add tzdata ca-certificates

COPY --from=builder /app/bot .

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"