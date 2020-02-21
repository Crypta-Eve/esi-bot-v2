  
FROM golang:1.13.8 as builder
WORKDIR /app
COPY . .
WORKDIR /app/cmd/eb2
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM alpine:latest AS release
WORKDIR /app/logs
WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates=20191127-r1

COPY --from=builder /app/cmd/eb2/eb2 .

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"