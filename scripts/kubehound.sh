#!/bin/bash

KUBEHOUND_ENV="release"
DOCKER_COMPOSE_FILE_PATH="-f deployments/kubehound/docker-compose.yaml"
DOCKER_COMPOSE_FILE_PATH+=" -f deployments/kubehound/docker-compose.release.yaml"
if [ -n "${DD_API_KEY}" ]; then
    DOCKER_COMPOSE_FILE_PATH+=" -f deployments/kubehound/docker-compose.datadog.yaml"
fi

DOCKER_COMPOSE_PROFILE="--profile infra"

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

DOCKER_HOSTNAME=$(hostname)
if [ -z "${CI}" ]; then
    DOCKER_CMD="DOCKER_HOSTNAME=${DOCKER_HOSTNAME} ${DOCKER_CMD}"
fi

run() {
    # TODO run kubehound with config file
}

backend_down() {
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} rm -fvs
}

backend_up() {
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} up --force-recreate --build -d
}

backend_reset() {
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} rm -fvs
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} up --force-recreate --build -d
}

backend_reset_hard() {
    ${DOCKER_CMD} volume rm kubehound-${KUBEHOUND_ENV}_mongodb_data
    ${DOCKER_CMD} volume rm kubehound-${KUBEHOUND_ENV}_janusgraph_data
    backend_reset()
}

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
