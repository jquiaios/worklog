FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /worklog ./cmd/worklog

FROM alpine:3.20

RUN adduser -D -u 1000 wl
USER wl
WORKDIR /home/wl

COPY --from=builder /worklog /usr/local/bin/worklog

ENTRYPOINT ["worklog"]
