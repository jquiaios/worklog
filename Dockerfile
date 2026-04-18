FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download || true

COPY . .
RUN go mod tidy && \
    go build -o /worklog ./cmd/worklog

FROM alpine:3.20

COPY --from=builder /worklog /usr/local/bin/worklog

ENTRYPOINT ["worklog"]
