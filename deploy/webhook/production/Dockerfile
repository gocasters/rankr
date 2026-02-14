FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -ldflags="-s -w" -o /app/main ./cmd/webhook/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/go.mod ./go.mod
COPY --from=builder /app/webhookapp/repository/migrations ./webhookapp/repository/migrations
COPY deploy/webhook/production/config.yml ./deploy/webhook/production/config.yml

EXPOSE 8090

CMD ["/app/main", "serve", "--migrate-up"]