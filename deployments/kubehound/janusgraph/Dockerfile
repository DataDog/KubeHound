FROM janusgraph/janusgraph:0.6.3
LABEL org.opencontainers.image.source="https://github.com/DataDog/kubehound/"

# Add our initialization script for the database schema to the startup directory
# See https://github.com/JanusGraph/janusgraph-docker#initialization
COPY kubehound-db-init.groovy /docker-entrypoint-initdb.d/

# Custom health check script
COPY health-check.groovy /opt/janusgraph/scripts/