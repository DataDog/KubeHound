# Build the KubeHound DSL .jar in a separate container
# maven:3-openjdk-11-slim
FROM maven@sha256:2cb7c73ba2fd0f7ae64cfabd99180030ec85841a1197b4ae821d21836cb0aa3b AS build

COPY dsl/kubehound/src /home/app/src
COPY dsl/kubehound/pom.xml /home/app

RUN mvn -f /home/app/pom.xml clean install

# Now build our janusgraph wrapper container with KubeHound customizations
# janusgraph/janusgraph:1.1.0
FROM janusgraph/janusgraph@sha256:75f57aff4b152ca86b4cdaeefb5335154d7fad412cffef75fa44c7504adc96d4
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
COPY --chown=janusgraph:janusgraph scripts/kubehound-dsl-init.groovy ${JANUS_HOME}/scripts/

# grpcurl to rehydrate dumps on startup
# fullstorydev/grpcurl:v1.9.1
COPY --from=fullstorydev/grpcurl@sha256:8bc96d11c8c08388b30cffafd177a1083a84e18c5bed314de5520c81171236a9 --chown=janusgraph:janusgraph /bin/grpcurl /usr/local/bin/grpcurl

# Set JVM configuration
ENV JAVA_OPTIONS_FILE ${JANUS_HOME}/conf/jvm.options

# Use an in-memory backend for speed
ENV JANUS_PROPS_TEMPLATE=inmemory

# Optimize for writes
ENV janusgraph.ids.block-size=1000000000
ENV janusgraph.ids.renew-percentage 0.3
ENV janusgraph.storage.batch-loading=true

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

# Add safety net settings to prevent OOM and other issues
ENV janusgraph.query.force-index=false
ENV janusgraph.cluster.max-partitions=512
ENV janusgraph.query.batch=true
ENV janusgraph.query.hard-max-limit=100000
ENV janusgraph.query.smart-limit=true

# Performance tweaks based on: https://www.sailpoint.com/blog/souping-up-the-gremlin/
# gremlinPool will default to Runtime.availableProcessors()
ENV gremlinserver.gremlinPool=0
# threadPoolWorker should be 2x VCPU (TODO: can we set dynamically?)
ENV gremlinserver.threadPoolWorker=16

# Custom SCRIPT plugin for DSL support
ENV gremlinserver.scriptEngines.gremlin-groovy.plugins[org.apache.tinkerpop.gremlin.jsr223.ImportGremlinPlugin].classImports[+]=com.datadog.ase.kubehound.KubeHoundTraversalSource
ENV gremlinserver.scriptEngines.gremlin-groovy.plugins[org.apache.tinkerpop.gremlin.jsr223.ImportGremlinPlugin].classImports[+]=com.datadog.ase.kubehound.EndpointExposure
ENV gremlinserver.scriptEngines.gremlin-groovy.plugins[org.apache.tinkerpop.gremlin.jsr223.ScriptFileGremlinPlugin].files[+]=scripts/kubehound-dsl-init.groovy

# Support for HTTP and WebSocket ports
ENV gremlinserver.channelizer=org.apache.tinkerpop.gremlin.server.channel.WsAndHttpChannelizer
