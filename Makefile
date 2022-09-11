build:
	go build -o bin/server ./cmd/server

test:
	go test -v ./... -coverprofile=coverage.out

cover:
	go tool cover -html=coverage.out

run:
	go run ./cmd/server

docker-build:
	docker build -t ghcr.io/dir01/mediary:alpha .

docker-push:
	docker push ghcr.io/dir01/mediary:alpha
