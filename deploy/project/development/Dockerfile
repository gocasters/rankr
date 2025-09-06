ARG GO_IMAGE_NAME
ARG GO_IMAGE_VERSION

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION}
ENV GOPROXY=https://goproxy.io,direct

# Set working directory
WORKDIR /home/app

RUN go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", "/home/app/.air/.air.user.toml"]