build:
	go build -o bin/service .

docker-build:
	docker build -t registry.gitlab.com/undercast/media_service .

docker-push:
	docker push registry.gitlab.com/undercast/media_service
