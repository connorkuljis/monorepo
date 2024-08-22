.PHONY: all server cli clean build push deploy

# Environment variables
GEMINIAPIKEY ?= $(shell echo $$GEMINIAPIKEY)

all: server cli

server:
	go build -v -o server cmd/server/***

cli:
	go build -v -o cli cmd/cli/***

clean:
	rm -f server cli

build:
	docker buildx build --platform linux/amd64 -t connorkuljis/seek:latest .

push: build
	docker push connorkuljis/seek:latest

deploy: push
	gcloud run deploy seek-test-1 --image 'docker.io/connorkuljis/seek' --allow-unauthenticated --set-env-vars=GEMINIAPIKEY="$(GEMINIAPIKEY)"
