#!/usr/bin/env sh

TAG=${PODLIKE_VERSION:-latest}

print_deploy_help() {
    echo """
Usage:	podtemplate deploy [OPTIONS] STACK

Loads the given templated YAML files, transforms them,
then deploys a new stack or updates an existing stack.

Options:
  -c, --compose-file strings   Path to a Compose file
      --prune                  Prune services that are no longer referenced
      --resolve-image string   Query the registry to resolve image digest and supported platforms (\"always\"|\"changed\"|\"never\") (default \"always\")
      --with-registry-auth     Send registry authentication details to Swarm agents
    """
}

exec_deploy() {
    COMPOSE_FILES=""
    STACK_DEPLOY_ARGS=""
    while [ $# -gt 0 ]
    do
    key="$1"

    case $key in
        -h|--help)
        print_deploy_help
        exit 0
        ;;
        -c|--compose-file)
        COMPOSE_FILES="$COMPOSE_FILES $2"
        shift
        shift
        ;;
        --resolve-image)
        STACK_DEPLOY_ARGS="$STACK_DEPLOY_ARGS $key $2"
        shift
        shift
        ;;
        --prune|--with-registry-auth)
        STACK_DEPLOY_ARGS="$STACK_DEPLOY_ARGS $key"
        shift
        ;;
        --bundle-file)
        echo 'Deploying Bundle files is not supported'
        exit 1
        ;;
        *)
        STACK_DEPLOY_ARGS="$STACK_DEPLOY_ARGS $key"
        shift
        ;;
    esac
    done

    if [ "$COMPOSE_FILES" = "" ]; then
        print_deploy_help
        exit 1
    fi

    # make sure the target image exist
    docker image inspect -f . rycus86/podlike:${TAG} 2>/dev/null >/dev/null || docker pull rycus86/podlike:${TAG}

    # generate the YAML ouput from the templates
    CONVERTED=$(docker run --rm -i -v $PWD:/workspace:ro -w /workspace rycus86/podlike:${TAG} template $COMPOSE_FILES 2>&1)
    if [ "$?" != "0" ]; then
        echo "$CONVERTED"
        exit 1
    fi

    # do the actual stack deployment
    echo """$CONVERTED""" | docker stack deploy -c - $STACK_DEPLOY_ARGS
}

exec_print() {
    # make sure the target image exist
    docker image inspect -f . rycus86/podlike:${TAG} 2>/dev/null >/dev/null || docker pull rycus86/podlike:${TAG}

    # generate the YAML ouput from the templates
    docker run --rm -i          \
        -v $PWD:/workspace:ro   \
        -w /workspace           \
        rycus86/podlike:${TAG}  \
        template $@   2>&1
}

# podtemplate deploy ...
if [ "$1" = "deploy" ]; then
    shift  # drop the deploy parameter
    exec_deploy $@

# podtemplate template ...
elif [ "$1" = "template" ]; then
    shift  # drop the deploy parameter
    exec_print $@

# podtemplate ...
else
    exec_print $@

fi
