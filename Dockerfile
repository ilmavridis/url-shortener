## Stage 1 - build
FROM golang:1.18.2-alpine3.16 AS builder

WORKDIR /build

COPY . .

RUN go mod download
RUN go build cmd/main.go


## Stage 2 - run
FROM alpine:3.16

LABEL maintainer='Ilias Mavridis'

RUN adduser -S -D -H -u 12222 -h /app appuser
USER appuser

WORKDIR /app

COPY --from=builder /build/main .
COPY .env ../
COPY config-default.yaml .
COPY webpage .

EXPOSE 80

ENTRYPOINT ["./main"]