FROM janusgraph/janusgraph:1.0.0-rc2
LABEL org.opencontainers.image.source="https://github.com/DataDog/kubehound/"

ENV JAVA_OPTIONS_FILE ${JANUS_HOME}/conf/jvm.options
# Add our initialization script for the database schema to the startup directory
# See https://github.com/JanusGraph/janusgraph-docker#initialization
COPY --chown=janusgraph:janusgraph kubehound-db-init.groovy /docker-entrypoint-initdb.d/

# Janusgraph metrics
COPY --chown=janusgraph:janusgraph lib/jmx_prometheus_javaagent-0.18.0.jar ${JANUS_HOME}/lib/jmx_prometheus_javaagent-0.18.0.jar
COPY --chown=janusgraph:janusgraph lib/exporter-config.yaml ${JANUS_HOME}/lib/exporter-config.yaml
COPY --chown=janusgraph:janusgraph conf/jvm.options ${JANUS_HOME}/conf/jvm.options
COPY --chown=janusgraph:janusgraph lib/kubehound-1.0.0.jar ${JANUS_HOME}/lib/kubehound-1.0.0.jar

# Custom health check script
COPY --chown=janusgraph:janusgraph scripts/health-check.groovy ${JANUS_HOME}/scripts/

# DSL support
COPY --chown=janusgraph:janusgraph kubehound-dsl-init.groovy ${JANUS_HOME}/scripts/
