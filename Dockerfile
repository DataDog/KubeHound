# syntax=docker/dockerfile:1

ARG GO_VERSION=1.22.0
ARG XX_VERSION=1.2.1
ARG GOLANGCI_LINT_VERSION=v1.55.2

# xx is a helper for cross-compilation
FROM --platform=${BUILDPLATFORM} tonistiigi/xx:${XX_VERSION} AS xx

# osxcross contains the MacOSX cross toolchain for xx
FROM crazymax/osxcross:11.3-alpine AS osxcross

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine AS base
COPY --from=xx / /
RUN apk add --no-cache \
    clang \
    docker \
    file \
    findutils \
    git \
    make \
    protoc \
    protobuf-dev
WORKDIR /src
ENV CGO_ENABLED=0

FROM base AS build-base

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY Makefile .
COPY cmd ./cmd
COPY configs ./configs
COPY deployments ./deployments
COPY .golangci.yml .golangci.yml

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

FROM build-base AS vendored
RUN --mount=type=bind,target=.,rw \
    --mount=type=cache,target=/go/pkg/mod \
    go mod tidy && mkdir /out && cp go.mod go.sum /out

FROM build-base AS build
ARG BUILD_TAGS
ARG BUILD_BRANCH
ARG BUILD_FLAGS
ARG TARGETPLATFORM
ENV BUILD_BRANCH="${BUILD_BRANCH}"
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,from=osxcross,src=/osxsdk,target=/xx-sdk \
    xx-go --wrap && \
    # Removing DWARD symbol on Darwin as it causes the following error:
    # /usr/local/go/pkg/tool/linux_arm64/link: /usr/local/go/pkg/tool/linux_arm64/link: running dsymutil failed: exec: "dsymutil": executable file not found in $PATH
    if [ "$(xx-info os)" == "darwin" ]; then export CGO_ENABLED=1; export BUILD_TAGS="-w $BUILD_TAGS"; fi && \
    make build GO_BUILDTAGS="$BUILD_TAGS" DESTDIR=/out && \
    xx-verify --static /out/kubehound

FROM build-base AS lint
ARG BUILD_TAGS
ENV GOLANGCI_LINT_CACHE=/cache/golangci-lint
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/cache/golangci-lint \
    --mount=from=golangci-lint,source=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
    golangci-lint cache status && \
    find / -iname .golangci.yaml && \
    pwd && \
    find / -iname .golangci.yaml -exec cat {} \; && \
    golangci-lint run --build-tags "$BUILD_TAGS" -c .golangci.yml ./...

FROM scratch AS binary-unix
COPY --link --from=build /out/kubehound /

FROM binary-unix AS binary-darwin
FROM binary-unix AS binary-linux

FROM scratch AS binary-windows
COPY --link --from=build /out/kubehound /kubehound.exe

FROM binary-$TARGETOS AS binary
# enable scanning for this stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true


FROM --platform=$BUILDPLATFORM alpine AS releaser
WORKDIR /work
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN --mount=from=binary \
    mkdir -p /out && \
    # TODO: should just use standard arch
    TARGETARCH=$([ "$TARGETARCH" = "amd64" ] && echo "x86_64" || echo "$TARGETARCH"); \
    # Use arm64 for darwin
    TARGETARCH=$([ "$TARGETARCH" = "arm64" ] && [ "$TARGETOS" != "darwin" ] && echo "aarch64" || echo "$TARGETARCH"); \
    # Upper case first letter to match the uname -o output
    TARGETOS=$([ "$TARGETOS" = "darwin" ] && echo "Darwin" || echo "$TARGETOS"); \
    TARGETOS=$([ "$TARGETOS" = "linux" ] && echo "Linux" || echo "$TARGETOS"); \
    cp kubehound* "/out/kubehound-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}$(ls kubehound* | sed -e 's/^kubehound//')"

FROM scratch AS release
COPY --from=releaser /out/ /
