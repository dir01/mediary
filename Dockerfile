# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app
# modernc.org/sqlite is pure Go, so we build without CGO and skip the C toolchain.
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ffmpeg python3

# Pin to a specific release (e.g. .../releases/download/2025.01.01/yt-dlp) for reproducible builds.
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp \
    && chmod +x /usr/local/bin/yt-dlp

COPY --from=builder /bin/server /usr/local/bin/server

CMD ["server"]
