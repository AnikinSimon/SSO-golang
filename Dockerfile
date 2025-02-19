FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN --mount=type=cache,target="./root/.cache/go-build" go build -o /app/sso /app/cmd/sso/

FROM ubuntu:24.04

WORKDIR /app
COPY --from=builder /app/sso .
COPY --from=builder /app/config/ /app/config/
ENTRYPOINT [ "/app/sso" ]