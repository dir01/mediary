FROM golang:1.25-alpine

RUN apk add --no-cache gcc musl-dev g++ ffmpeg python3

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && chmod +x /usr/local/bin/yt-dlp

RUN mkdir /app
ENV GOPATH ""
ADD go.mod go.sum ./
RUN go mod download

ADD . .
RUN GOPATH= go build -o bin/server ./cmd/server

CMD bin/server
