FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache netcat-openbsd

# Install air for hot-reload
RUN go install github.com/air-verse/air@latest

ENV GOPROXY=https://mirrors.aliyun.com/goproxy,direct

COPY go.mod go.sum ./
# RUN go mod download

COPY . .

EXPOSE 8080 8081 8091

# Use air as the entrypoint (will be overridden by docker-compose)
CMD ["air", "-c", "/app/.air/.air.leaderboardstat.toml"]