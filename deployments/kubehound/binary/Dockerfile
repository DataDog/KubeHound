FROM golang:1.22 AS build-stage

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY Makefile .
COPY cmd ./cmd
COPY configs ./configs
COPY deployments ./deployments

RUN make build

FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /go/bin/build/kubehound /kubehound

EXPOSE 9000

USER nonroot:nonroot

ENTRYPOINT ["/kubehound"]
CMD  ["serve"]