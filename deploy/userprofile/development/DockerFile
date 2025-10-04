# ===== Base Image =====
ARG GO_IMAGE_NAME
ARG GO_IMAGE_VERSION
FROM ${GO_IMAGE_NAME}:${GO_IMAGE_VERSION}

# ===== Environment =====
ENV GOPROXY=https://goproxy.io,direct
ENV PATH=$PATH:/go/bin

# ===== Working Directory =====
WORKDIR /home/app

# ===== Install Air =====
RUN go install github.com/air-verse/air@latest

# ===== Dependencies =====
COPY go.mod go.sum ./
RUN go mod download

# ===== Copy Source Code =====
COPY . .

# ===== CMD: Start Hot Reload =====
CMD ["air", "-c", "/home/app/.air/.air.userprofile.toml"]

