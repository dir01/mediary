build: prebuild
	@echo "+ $@"
	go build -o bin/server ./cmd/server

.PHONY: run
run:
	@echo "+ $@"
	REDIS_URL=redis://localhost:6379 go run ./cmd/server

.PHONY: test
test:
	@echo "+ $@"
	go test -v -failfast -race ./... -coverprofile=coverage.out

.PHONY: prebuild
prebuild:
	@echo "+ $@"
	go mod tidy
	go mod vendor

.PHONY: precommit
precommit: prebuild lint test

lint:
	docker run -t --rm -v $$(pwd):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v --timeout 5m

.PHONY: test-e2e-gen-docs
test-e2e-gen-docs:
	@echo "+ $@"
	go test -v -timeout 30m -failfast -race -tags gen_docs ./... -coverprofile=coverage.out

.PHONY: generate
generate:
	@echo "+ $@"
	go generate ./...

.PHONY: cover
cover:
	@echo "+ $@"
	go tool cover -html=coverage.out

.PHONY: docker-build
docker-build:
	@echo "+ $@"
	docker build -t ghcr.io/dir01/mediary:alpha .

.PHONY: docker-push
docker-push:
	@echo "+ $@"
	docker push ghcr.io/dir01/mediary:alpha
