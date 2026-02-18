# syntax=docker/dockerfile:1
FROM golang:1.25-trixie AS build-stage

WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/nosmallparts/backend

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY vendor vendor
COPY go.mod go.mod
COPY go.sum go.sum

RUN --mount=type=cache,target=/root/.cache/go-build go build -trimpath -o /nosmallparts github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/api/http

# final stage
FROM debian:bullseye

# RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
COPY --from=build-stage /nosmallparts /nosmallparts

ENTRYPOINT ["/nosmallparts"]
