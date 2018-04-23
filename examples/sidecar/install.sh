#!/usr/bin/env sh

set -x

git clone https://github.com/rycus86/podlike.git
cd podlike/examples/sidecar
docker stack deploy -c stack.yml sidecar
