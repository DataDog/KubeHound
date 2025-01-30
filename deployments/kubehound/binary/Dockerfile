# golang:1.23
FROM golang@sha256:8c10f21bec412f08f73aa7b97ca5ac5f28a39d8a88030ad8a339fd0a781d72b4 AS build-stage

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY Makefile .
COPY cmd ./cmd
COPY configs ./configs
COPY deployments ./deployments

RUN make build

# gcr.io/distroless/base-debian12
FROM gcr.io/distroless/base-debian12@sha256:74ddbf52d93fafbdd21b399271b0b4aac1babf8fa98cab59e5692e01169a1348 AS build-release-stage

WORKDIR /

COPY --from=build-stage /go/bin/build/kubehound /kubehound

EXPOSE 9000

USER nonroot:nonroot

ENTRYPOINT ["/kubehound"]
CMD  ["serve"]
