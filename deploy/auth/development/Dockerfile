FROM golang:1.25-alpine AS builder
ENV CGO_ENABLED=0 \
    GOOS=linux
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags "-s -w" -o /auth ./authapp/app.go

# Use a minimal, empty base image for the final runtime
FROM scratch

ENV TZ=UTC

COPY --from=builder /auth /auth

EXPOSE 5002

ENTRYPOINT ["/auth"]
