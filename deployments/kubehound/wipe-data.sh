#!/bin/bash
set -e

rm -r ./data/mongodb/
rm -r ./data/neo4j/

# JanusGraph data currently held in memory. Wipe via docker restart