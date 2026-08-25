FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /worklog ./cmd/worklog

FROM alpine:3.24 AS app

RUN adduser -D -u 1000 wl
USER wl
WORKDIR /home/wl

COPY --from=builder /worklog /usr/local/bin/worklog

ENTRYPOINT ["worklog"]

FROM ghcr.io/charmbracelet/vhs AS vhs

COPY --from=builder /worklog /usr/local/bin/worklog
