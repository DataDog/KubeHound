# gcr.io/datadoghq/agent:7-jmx
FROM gcr.io/datadoghq/agent@sha256:465da828947d6c055eacf399df8d0f96423f6e029fd1218a806f091a0dc0200b

ADD datadog.yaml /etc/datadog-agent/datadog.yaml
ADD mongodb.yaml /etc/datadog-agent/conf.d/mongo.d/conf.yaml
ADD openmetrics.yaml /etc/datadog-agent/conf.d/openmetrics.d/conf.yaml
