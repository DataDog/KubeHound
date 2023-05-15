#!/bin/bash
set -e

# Spin up the compose with neo4j, mongodb, etc
docker-compose up --force-recreate