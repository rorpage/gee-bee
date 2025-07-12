# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o gee-bee ./cmd/geebee/*

# Run stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/gee-bee .

ENV CUSTOM_WEBHOOK_URL=""
ENV DISCORD_WEBHOOK_URL=""
ENV FETCH_INTERVAL=60
ENV LOG_PLANES_TO_CONSOLE=true
ENV SLACK_WEBHOOK_URL=""
ENV TAIL_NUMBERS="28000,29000"

CMD ["./gee-bee"]
