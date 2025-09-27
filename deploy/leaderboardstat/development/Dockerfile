ARG GO_IMAGE_NAME=golang
ARG GO_IMAGE_VERSION=1.25.0-alpine

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION} AS builder

# Set GOPROXY with direct fallback
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

WORKDIR /home/app

# Copy from project root (context is set to ../../.. in compose)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/leaderboardstat cmd/leaderboardstat/main.go

FROM alpine:3.19

RUN echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/main" > /etc/apk/repositories && \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/community" >> /etc/apk/repositories

COPY --from=builder /bin/leaderboardstat /bin/leaderboardstat

WORKDIR /home/app

EXPOSE 6011

CMD ["/bin/leaderboardstat", "serve"]