build:
	@echo "+ $@"
	go build -o bin/server ./cmd/server

.PHONY: test
test:
	@echo "+ $@"
	go test -v ./... -coverprofile=coverage.out

.PHONY: test-e2e-gen-docs
test-e2e-gen-docs:
	@echo "+ $@"
	go test -v -timeout 1m -failfast -race -tags gen_docs -parallel 1 ./tests

.PHONY: cover
cover:
	@echo "+ $@"
	go tool cover -html=coverage.out

.PHONY: run
run:
	@echo "+ $@"
	go run ./cmd/server

.PHONY: docker-build
docker-build:
	@echo "+ $@"
	docker build -t ghcr.io/dir01/mediary:alpha .

.PHONY: docker-push
docker-push:
	@echo "+ $@"
	docker push ghcr.io/dir01/mediary:alpha
