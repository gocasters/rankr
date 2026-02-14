FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -ldflags="-s -w" -o /app/main ./cmd/notification/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/go.mod ./go.mod
COPY --from=builder /app/notifapp/repository/migrations ./notifapp/repository/migrations
COPY deploy/notification/production/config.yml ./deploy/notification/production/config.yml

EXPOSE 8080

CMD ["/app/main", "serve", "--migrate-up"]