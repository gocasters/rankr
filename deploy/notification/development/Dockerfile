FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/notification/main.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/notifapp/repository/migrations ./notifapp/repository/migrations

EXPOSE 8080

CMD ["/app/main", "serve"]