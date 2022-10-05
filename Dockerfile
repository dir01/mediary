FROM golang:alpine
RUN apk add gcc musl-dev g++
ENV GOPATH ""
ADD go.mod go.sum ./
RUN go mod download
ADD . .
RUN GOPATH= go build -o bin/server ./cmd/server
CMD bin/service
