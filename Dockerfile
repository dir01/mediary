FROM golang:1.26-alpine

RUN apk add --no-cache gcc musl-dev g++ ffmpeg python3

ARG TARGETARCH
RUN case "${TARGETARCH}" in \
    "amd64") YTDLP_BINARY="yt-dlp" ;; \
    "arm64") YTDLP_BINARY="yt-dlp_linux_aarch64" ;; \
    *) echo "Unsupported architecture: ${TARGETARCH}" && exit 1 ;; \
  esac && \
  wget "https://github.com/yt-dlp/yt-dlp/releases/latest/download/${YTDLP_BINARY}" -O /usr/local/bin/yt-dlp && \
  chmod +x /usr/local/bin/yt-dlp

WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download

ADD . .
RUN go build -o bin/server ./cmd/server

CMD ["/app/bin/server"]
