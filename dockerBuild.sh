
#! /bin/sh.
DOCKER_BUILDKIT=0 docker build --rm . -t crimsonq:latest
docker image prune --filter label=stage=builder
