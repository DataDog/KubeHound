#!/bin/bash
set -e

# Spin up the compose with janusgraph, mongodb, datadog, etc
docker-compose -f ./docker-compose-dev.yaml up --force-recreate --build