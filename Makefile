build:
	@echo "+ $@"
	go build -o bin/server ./cmd/server
.PHONY: build

run:
	@echo "+ $@"
	REDIS_URL=redis://localhost:6379 go run ./cmd/server
.PHONY: run

test:
	@echo "+ $@"
	go test -v -failfast -race ./... -coverprofile=coverage.out
.PHONY: test

tidy:
	@echo "+ $@"
	go mod tidy
.PHONY: tidy

precommit: tidy build lint test
.PHONY: precommit

lint:
	docker run -t --rm -v $$(pwd):/app -w /app golangci/golangci-lint:v2.0.2 golangci-lint run -v --timeout 5m
.PHONY: lint

test-e2e-gen-docs:
	@echo "+ $@"
	go test -v -timeout 30m -failfast -race -tags gen_docs ./... -coverprofile=coverage.out
.PHONY: test-e2e-gen-docs

generate: prebuild
	@echo "+ $@"
	go generate ./...
.PHONY: generate

cover:
	@echo "+ $@"
	go tool cover -html=coverage.out
.PHONY: cover

docker-build:
	@echo "+ $@"
	docker build -t ghcr.io/dir01/mediary:alpha .
.PHONY: docker-build

docker-push:
	@echo "+ $@"
	docker push ghcr.io/dir01/mediary:alpha
.PHONY: docker-push
