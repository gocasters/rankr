ARG golang
ARG 1.23.4-alpine

FROM golang:1.23.4-alpine
ENV GOPROXY=https://goproxy.io,direct


# Set working directory
WORKDIR /home/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .,
