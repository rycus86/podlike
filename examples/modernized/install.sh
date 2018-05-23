#!/usr/bin/env sh

set -x

git clone https://github.com/rycus86/podlike.git
cd podlike/examples/modernized
docker stack deploy -c stack.yml modern
