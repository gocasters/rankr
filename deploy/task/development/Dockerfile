FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/air-verse/air@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/task/main.go


FROM golang:1.25-alpine

WORKDIR /app

COPY --from=builder /go/bin/air /usr/local/bin/air
COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /go/pkg/sumdb /go/pkg/sumdb
COPY --from=builder /app /app

EXPOSE 8080

CMD ["/app/main", "serve", "--migrate-up"]
