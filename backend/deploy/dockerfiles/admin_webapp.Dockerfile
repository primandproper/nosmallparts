# build stage
FROM golang:1.25-trixie AS build-stage

WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/nosmallparts/backend

COPY . .

RUN go build -trimpath -o /server github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin

# final stage
FROM debian:bullseye

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
COPY --from=build-stage /server /server

ENTRYPOINT ["/server"]
