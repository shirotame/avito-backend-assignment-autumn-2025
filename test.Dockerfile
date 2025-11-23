FROM golang:1.25.3-alpine

WORKDIR app

COPY internal ./internal/
COPY migrations ./migrations/
COPY mocks ./mocks/
COPY test ./test/
COPY go.mod go.sum ./