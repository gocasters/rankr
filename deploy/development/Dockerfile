ARG GO_IMAGE_NAME=golang
ARG GO_IMAGE_VERSION=1.25.3-alpine
ARG AIR_VERSION=latest
ARG GOPROXY=https://mirrors.aliyun.com/goproxy,direct

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION}
ARG AIR_VERSION

WORKDIR /app
ENV GOPROXY=${GOPROXY}
RUN go install github.com/air-verse/air@${AIR_VERSION}

COPY go.mod go.sum ./
RUN go mod download