#!/usr/bin/env bash
docker buildx build $DOCKER_ARGS \
  --build-arg CADDY_VERSION=$(grep caddy .tool-versions | awk '{print $2}') \
  .
