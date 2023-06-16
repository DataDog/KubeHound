FROM janusgraph/janusgraph:latest

# Add our initialization script for the database schema to the startup directory
# See https://github.com/JanusGraph/janusgraph-docker#initialization
COPY kubehound-db-init.groovy /docker-entrypoint-initdb.d/

# 
COPY health-check.groovy /opt/janusgraph/scripts/