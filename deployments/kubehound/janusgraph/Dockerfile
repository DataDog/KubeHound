FROM janusgraph/janusgraph:1.0.0-rc2
LABEL org.opencontainers.image.source="https://github.com/DataDog/kubehound/"

ENV JAVA_OPTIONS_FILE /opt/janusgraph/conf/jvm.options
# Add our initialization script for the database schema to the startup directory
# See https://github.com/JanusGraph/janusgraph-docker#initialization
COPY kubehound-db-init.groovy /docker-entrypoint-initdb.d/

# Janusgraph metrics
COPY --chown=janusgraph:janusgraph lib/jmx_prometheus_javaagent-0.18.0.jar ${JANUS_HOME}/lib/jmx_prometheus_javaagent-0.18.0.jar
COPY --chown=janusgraph:janusgraph lib/exporter-config.yaml ${JANUS_HOME}/lib/exporter-config.yaml
COPY --chown=janusgraph:janusgraph conf/jvm.options ${JANUS_HOME}/conf/jvm.options

# Specific config file for Kubehound needs
COPY --chown=janusgraph:janusgraph conf/gremlin-server.yaml ${JANUS_HOME}/conf/

# Custom health check script
COPY --chown=janusgraph:janusgraph scripts/health-check.groovy ${JANUS_HOME}/scripts/