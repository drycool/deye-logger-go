# Stage 1: Build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo 'dev') -X main.commit=$(git log -1 --format=%h) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /deye-logger ./cmd/deye-logger/

# Stage 2: Runtime
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /deye-logger /deye-logger
COPY config/deye.example.yml /config/deye.example.yml

EXPOSE 8899

ENTRYPOINT ["/deye-logger"]
CMD ["--config", "/config/deye.yml"]
