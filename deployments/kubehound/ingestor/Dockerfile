FROM golang:1.22 AS build-stage

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY Makefile .
COPY cmd ./cmd
COPY configs ./configs

RUN make build-ingestor

FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /go/bin/kubehound-ingestor /kubehound-ingestor

EXPOSE 9000

USER nonroot:nonroot

ENTRYPOINT ["./kubehound-ingestor"]