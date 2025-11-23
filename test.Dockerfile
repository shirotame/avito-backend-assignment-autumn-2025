FROM golang:1.25.3-alpine

WORKDIR app

COPY internal ./internal/
COPY migrations ./migrations/
COPY test ./test/
COPY go.mod go.sum ./