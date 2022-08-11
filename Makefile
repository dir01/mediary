build:
	go build -o bin/service .

test:
	go test -v ./...

run:
	go run ./cmd/server

docker-build:
	docker build -t ghcr.io/dir01/mediary:alpha .

docker-push:
	docker push ghcr.io/dir01/mediary:alpha
