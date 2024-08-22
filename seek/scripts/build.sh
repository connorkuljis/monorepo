#!/bin/zsh
set -e 
set -x

docker buildx build --platform linux/amd64 -t connorkuljis/seek:latest .
docker push connorkuljis/seek:latest

gcloud run deploy seek-test-1 --image 'docker.io/connorkuljis/seek' --allow-unauthenticated --set-env-vars=GEMINIAPIKEY="$GEMINIAPIKEY"
