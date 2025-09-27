FROM golang:1.25-alpine

WORKDIR /home/app

# Install dependencies for benchmarking and debugging
RUN apk add --no-cache \
    postgresql-client \
    curl \
    htop \
    procps

# Copy go mod and sum files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Install go benchmarks tools
RUN go install golang.org/x/perf/cmd/benchstat@latest

# Default command (will be overridden by compose)
CMD ["sleep", "infinity"]