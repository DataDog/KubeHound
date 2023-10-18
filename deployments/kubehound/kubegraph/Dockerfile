# Build the KubeHound DSL .jar in a separate container
FROM maven:3-openjdk-11-slim AS build

COPY dsl/kubehound/src /home/app/src
COPY dsl/kubehound/pom.xml /home/app

RUN mvn -f /home/app/pom.xml clean install

# Now build our janusgraph wrapper container with KubeHound customizations
FROM janusgraph/janusgraph:1.0.0-rc2
LABEL org.opencontainers.image.source="https://github.com/DataDog/kubehound/"

# Add our initialization script for the database schema to the startup directory
# See https://github.com/JanusGraph/janusgraph-docker#initialization
COPY --chown=janusgraph:janusgraph kubehound-db-init.groovy /docker-entrypoint-initdb.d/

# Janusgraph metrics
COPY --chown=janusgraph:janusgraph lib/jmx_prometheus_javaagent-0.18.0.jar ${JANUS_HOME}/lib/jmx_prometheus_javaagent-0.18.0.jar
COPY --chown=janusgraph:janusgraph lib/exporter-config.yaml ${JANUS_HOME}/lib/exporter-config.yaml
COPY --chown=janusgraph:janusgraph conf/jvm.options ${JANUS_HOME}/conf/jvm.options

# Copy the DSL .jar from the build container
COPY --from=build --chown=janusgraph:janusgraph /home/app/target/kubehound-1.0.0.jar ${JANUS_HOME}/lib/kubehound-1.0.0.jar

# Custom health check script
COPY --chown=janusgraph:janusgraph scripts/health-check.groovy ${JANUS_HOME}/scripts/

# DSL support
COPY --chown=janusgraph:janusgraph kubehound-dsl-init.groovy ${JANUS_HOME}/scripts/

# Set JVM configuration
ENV JAVA_OPTIONS_FILE ${JANUS_HOME}/conf/jvm.options

# Use an in-memory backend for speed
ENV JANUS_PROPS_TEMPLATE=inmemory

# Optimize for writes
ENV janusgraph.ids.block-size=1000000000
ENV janusgraph.ids.renew-percentage 0.3
ENV storage.batch-loading=true

# Enforce strict schema constraints as per https://docs.janusgraph.org/configs/configuration-reference/#schema
ENV janusgraph.schema.constraints=true
ENV janusgraph.schema.default=none

# Bump content length of web-socket buffer to enable bulk insert queries
ENV gremlinserver.maxContentLength=2097152

# Bump evaluation timeout
ENV gremlinserver.evaluationTimeout=240000

# Enable metrics only for jmxReporter
ENV gremlinserver.metrics.jmxReporter.enabled=true
ENV gremlinserver.metrics.consoleReporter.enabled=false
ENV gremlinserver.metrics.slf4jReporter.enabled=false
ENV gremlinserver.metrics.graphiteReporter.enabled=false
ENV gremlinserver.metrics.csvReporter.enabled=false

# Performance tweaks based on: https://www.sailpoint.com/blog/souping-up-the-gremlin/
# gremlinPool will default to Runtime.availableProcessors()
ENV gremlinserver.gremlinPool=0 
# threadPoolWorker should be 2x VCPU (TODO: can we set dynamically?)
ENV gremlinserver.threadPoolWorker=16

# Custom SCRIPT plugin for DSL support
ENV gremlinserver.scriptEngines.gremlin-groovy.plugins[org.apache.tinkerpop.gremlin.jsr223.ImportGremlinPlugin].classImports[+]=com.datadog.ase.kubehound.KubeHoundTraversalSource
ENV gremlinserver.scriptEngines.gremlin-groovy.plugins[org.apache.tinkerpop.gremlin.jsr223.ImportGremlinPlugin].classImports[+]=com.datadog.ase.kubehound.EndpointExposure
ENV gremlinserver.scriptEngines.gremlin-groovy.plugins[org.apache.tinkerpop.gremlin.jsr223.ScriptFileGremlinPlugin].files[+]=scripts/kubehound-dsl-init.groovy
