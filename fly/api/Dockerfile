# build stage
FROM golang:1.25-trixie AS build-stage

WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/nosmallparts/backend

COPY backend/ .

RUN go build -trimpath -o /server github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/api

# final stage
FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /uploads /config

COPY fly/api/config.json /config/config.json
COPY --from=build-stage /server /server

ENTRYPOINT ["/server"]
