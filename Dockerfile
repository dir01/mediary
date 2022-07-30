FROM golang:alpine
COPY . .
RUN GOPATH= go build -o bin/service .
CMD bin/service
