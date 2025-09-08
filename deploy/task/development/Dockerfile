ARG GO_IMAGE_NAME
ARG GO_IMAGE_VERSION

FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION}
ENV GOPROXY=https://goproxy.io,direct

# Add /go/bin to PATH to make air accessible
ENV PATH=$PATH:/go/bin

# Set working directory
WORKDIR /home/app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# Install air
RUN go install github.com/air-verse/air@latest

# Verify air installation (optional, for debugging)
RUN air -v
