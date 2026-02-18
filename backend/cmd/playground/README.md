# Playground

This is a place to use the available API for experiments.

## Bootstrap scripts (main.go)

Loads bootstrap play scripts (movies, TV) and inserts them directly into the database. Requires a database connection (e.g. via Cloud SQL proxy or port-forward).

```bash
# 1. Proxy the database (e.g. Cloud SQL Auth Proxy or kubectl port-forward)
# make proxy_dev_db   # or your preferred method

# 2. Set env vars
export PLAYGROUND_DATABASE_URL="postgres://user:password@127.0.0.1:5434/no-small-parts?sslmode=disable"
# Or use individual vars:
# export PLAYGROUND_DB_HOST=127.0.0.1
# export PLAYGROUND_DB_PORT=5434
# export PLAYGROUND_DB_NAME=no-small-parts
# export PLAYGROUND_DB_USER=...
# export PLAYGROUND_DB_PASSWORD=...

# 3. Ownership - use valid user/account IDs from the target DB
export PLAYGROUND_CREATED_BY_USER=<user_id>
export PLAYGROUND_BELONGS_TO_ACCOUNT=<account_id>

# 4. Run
go run ./cmd/playground
```

Do not commit real credentials.

## Connecting to the dev database

To connect to dev, you'll need to run `make proxy_dev_db`, and then run code that connects to the database.

Since all Go files in this folder aren't saved, here's a handy template for what the database connection string needs to look like:

```go
const dbString = `user=<user> password=<password> database=<database> host=127.0.0.1 port=5434 sslmode=disable`
```

Do not save real credentials to this file or anything that would end up in source control.

### Examples

## using the API client

```go
package main

import (
 "context"
 "net/url"
 "time"

 "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/tracing"
 "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/apiclient"
 "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/types"

 "github.com/pquerna/otp/totp"
)

func main() {
 u, err := url.Parse("https://api.nosmall.parts")
 if err != nil {
  panic(err)
 }

 client, err := apiclient.NewClient(u, tracing.NewNoopTracerProvider())
 if err != nil {
  panic(err)
 }

 code, err := totp.GenerateCode("REPLACEMEREPLACEMEREPLACEMEREPLACEMEREPLACEMEREPLACEMEREPLACEMEREPLACEMEREPLACEMEREPLACEMEREPLACEMEYUSS", time.Now())
 if err != nil {
  panic(err)
 }

 jwtResponse, err := client.LoginForToken(context.Background(), &types.UserLoginInput{
  Username:  "username",
  Password:  "password",
  TOTPToken: code,
 })
 if err != nil {
  panic(err)
 }

 println(jwtResponse)
}
```

## using the DataManager

```go
package main

import (
 "context"
 "fmt"
 "os"
 "log"
 "time"
 
 databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/database/routingcfg"
 "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/database/postgres"
 "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/logging"
 "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/tracing"
)

const dbString = `user=%s password=%s database=%s host=127.0.0.1 port=5434 sslmode=disable`

func main() {
 ctx := context.Background()
 logger := logging.NewNoopLogger()
 tracerProvider := tracing.NewNoopTracerProvider()

 dbUser := os.Getenv("DEV_DATABASE_USER")
 dbPassword := os.Getenv("DEV_DATABASE_PASSWORD")
 dbName := os.Getenv("DEV_DATABASE_DB")

 databaseConfig := &databasecfg.Config{
  OAuth2TokenEncryptionKey: os.Getenv("DEV_DATABASE_OAUTH2_ENCRYPTION_KEY"),
  ConnectionDetails:        fmt.Sprintf(dbString, dbUser, dbPassword, dbName),
  RunMigrations:            false,
  MaxPingAttempts:          10,
  PingWaitPeriod:           time.Second,
 }

 dbConnectionContext, cancel := context.WithTimeout(ctx, 15*time.Second)
 dataManager, err := postgres.ProvideDatabaseClient(dbConnectionContext, logger, tracerProvider, databaseConfig)
 if err != nil {
  log.Fatal(err)
 }

 cancel()
 defer dataManager.Close()

 collection, err := dataManager.AggregateUserData(ctx, "cepmmrq23akg00b01aqg")
 if err != nil {
  log.Fatal(err)
 }

 log.Println(collection)
}
```
