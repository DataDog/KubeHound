#!/bin/bash

#
# Lightweight wrapper script to run KubeHound from a release archive
#

# Set the environment as the release environment
KUBEHOUND_ENV="release"

# Pull in the requisite compose files for the current setup
DOCKER_COMPOSE_FILE_PATH="-f deployments/kubehound/docker-compose.yaml"
DOCKER_COMPOSE_FILE_PATH+=" -f deployments/kubehound/docker-compose.release.yaml"
if [ -n "${DD_API_KEY}" ]; then
    DOCKER_COMPOSE_FILE_PATH+=" -f deployments/kubehound/docker-compose.datadog.yaml"
fi

# Set the environment variables for the compose
DOCKER_COMPOSE_PROFILE="--profile infra"
DOCKER_HOSTNAME=$(hostname)

# Resolve the correct docker command for this environment (Linux requires sudo)
UNAME_S=$(uname -s)
if [ -z "${DOCKER_CMD}" ]; then
    if [ "${UNAME_S}" == "Linux" ]; then
        if [ -z "${CI}" ]; then
            DOCKER_CMD="sudo docker"
        else
            DOCKER_CMD="docker"
        fi
    else
        DOCKER_CMD="docker"
    fi
    DOCKER_CMD="${DOCKER_CMD}"
fi

# Run the kubehound binary
run() {
    ./kubehound -c config.yaml
}

# Shut down the kubehound backend
backend_down() {
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} rm -fvs
}

# Bring up the kubehound backend
backend_up() {
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} up --force-recreate --build -d
}

# Reset the kubehound backend
backend_reset() {
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} rm -fvs
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} up --force-recreate --build -d
}

# Reset the kubehound backend (WIPING ALL DATA)
backend_reset_hard() {
    backend_down
    ${DOCKER_CMD} volume rm kubehound-${KUBEHOUND_ENV}_mongodb_data
    ${DOCKER_CMD} volume rm kubehound-${KUBEHOUND_ENV}_kubegraph_data
    backend_up
}

# Entrypoint
case "$1" in
    run)
        run
        ;;
    backend-down)
        backend_down
        ;;
    backend-up)
        backend_up
        ;;
    backend-reset)
        backend_reset
        ;;
    backend-reset-hard)
        backend_reset_hard
        ;;
    *)
        echo "Usage: $0 {run|backend-up|backend-reset|backend-reset-hard|backend-down}"
        exit 1
esac
