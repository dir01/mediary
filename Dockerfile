FROM golang:1.24-alpine

RUN apk add --no-cache gcc musl-dev g++ ffmpeg git python3

ENV YOUTUBEDL_DIR /opt/youtube-dl
ENV YOUTUBEDL_REV 4549522
RUN git clone https://github.com/ytdl-org/youtube-dl $YOUTUBEDL_DIR
RUN cd $YOUTUBEDL_DIR && git checkout $YOUTUBEDL_REV

RUN mkdir /app
ENV GOPATH ""
ADD go.mod go.sum ./
RUN go mod download

ADD . .
RUN GOPATH= go build -o bin/server ./cmd/server

CMD bin/server
