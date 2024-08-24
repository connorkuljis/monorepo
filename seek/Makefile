.PHONY: all server clean build push deploy

# Environment variables
GEMINIAPIKEY ?= $(shell echo $$GEMINIAPIKEY)

all: server cli

server:
	docker build -t connorkuljis/seek:dev .
	docker run -p 8080:8080 -e GEMINIAPIKEY="$(GEMINIAPIKEY)" -t connorkuljis/seek:dev

clean:
	docker image prune -a
	rm -f server cli

build:
	docker buildx build --platform linux/amd64 -t connorkuljis/seek:latest .

push: build
	docker push connorkuljis/seek:latest

deploy: push
	gcloud run deploy seek-test-1 --image 'docker.io/connorkuljis/seek' --allow-unauthenticated --set-env-vars=GEMINIAPIKEY="$(GEMINIAPIKEY)"
