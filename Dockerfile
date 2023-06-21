FROM golang:1.20-alpine AS build

RUN apk update && \
    apk add make && \
    rm -rf /var/cache/apt/* && \
    go install github.com/vektra/mockery/v2@v2.30.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY Makefile ./
COPY cmd ./cmd/
COPY pkg ./pkg/

RUN make build BUILD_VERSION=${VERSION}

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/DataDog/kubehound/"

WORKDIR /
COPY --from=build /app/bin/kubehound /kubehound
COPY deployments/kubehound/kubehound.yaml /etc/kubehound.yaml

ENTRYPOINT [ "/kubehound" ]