build:
	go build -o bin/service .

test:
	go test -v ./...

docker-build:
	docker build -t ghcr.io/dir01/mediary:alpha .

docker-push:
	docker push ghcr.io/dir01/mediary:alpha
