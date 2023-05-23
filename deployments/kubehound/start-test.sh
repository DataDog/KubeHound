#!/bin/bash
set -e

# Spin up the compose with neo4j, mongodb, etc
docker-compose -f ./docker-compose-test.yaml up --force-recreate