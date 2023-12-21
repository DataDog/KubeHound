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
export DOCKER_HOSTNAME

# Make sure we have docker installed and we can access it (group / sudo permission)
# This function is only called when docker is required
check_docker() {
    if [ -n "${DOCKER_CMD}" ]; then
        return
    fi

    if ! [ -x "$(command -v docker)" ]; then
    # docker isn't available at all, there's no point in continuing
        echo "Docker isn't available. You should install it."
        exit 1
    fi

    if ! [ "$(docker compose version | grep '^Docker Compose version 2')" ]; then
    # docker compose v2 isn't available at all, there's no point in continuing
        echo "Docker Compose v2 isn't available. You should install it."
        exit 1
    fi

    DOCKER_CMD="docker"
    if ! $DOCKER_CMD info > /dev/null 2>&1; then
        echo "Docker isn't accessible with the current user. Retrying with sudo."
        # We need to pass the env vars (DOCKER_HOSTNAME and DD_API_KEY) to sudo
        DOCKER_CMD="sudo DOCKER_HOSTNAME=${DOCKER_HOSTNAME} DD_API_KEY=${DD_API_KEY} docker"
    fi

    if ! $DOCKER_CMD info > /dev/null 2>&1; then
        echo "We don't have the permission to run docker. Are you root or in the docker group?"
        exit 1
    fi
}

# Run the kubehound binary
run() {
    ./kubehound -c config.yaml
}

# Shut down the kubehound backend
backend_down() {
    check_docker
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} rm -fvs
}

# Bring up the kubehound backend
backend_up() {
    check_docker
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} up --force-recreate --build -d
}

# Reset the kubehound backend
backend_reset() {
    check_docker
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} rm -fvs
    ${DOCKER_CMD} compose ${DOCKER_COMPOSE_FILE_PATH} ${DOCKER_COMPOSE_PROFILE} up --force-recreate --build -d
}

# Reset the kubehound backend (WIPING ALL DATA)
backend_reset_hard() {
    check_docker
    backend_down
    ${DOCKER_CMD} volume rm "kubehound-${KUBEHOUND_ENV}_mongodb_data"
    ${DOCKER_CMD} volume rm "kubehound-${KUBEHOUND_ENV}_kubegraph_data"
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
