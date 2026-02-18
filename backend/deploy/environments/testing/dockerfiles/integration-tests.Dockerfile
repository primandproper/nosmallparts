# syntax=docker/dockerfile:1
FROM golang:1.25-trixie

WORKDIR /go/src/github.com/verygoodsoftwarenotvirus/nosmallparts/backend
COPY . .

# to debug a specific test:
# ENTRYPOINT go test -parallel 1 -v -failfast github.com/verygoodsoftwarenotvirus/nosmallparts/backend/tests/integration -run TestIntegration/TestValidPreparationInstruments_CompleteLifecycle

ENTRYPOINT ["go", "test", "-v", "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/tests/integration"]
