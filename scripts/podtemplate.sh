#!/usr/bin/env sh

TAG=${PODLIKE_VERSION:-latest}

docker run --rm -i          \
    -v $PWD:/workspace:ro   \
    -w /workspace           \
    rycus86/podlike:${TAG}  \
    template $@
