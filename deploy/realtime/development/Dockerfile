ARG GO_IMAGE_NAME=golang
ARG GO_IMAGE_VERSION=1.25-alpine3.21

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION}

WORKDIR /home/app

# Install air for hot reloading and other tools
RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# The actual command will be provided by docker-compose
CMD ["air", "-c", ".air/.air.realtime.toml"]