package main

import (
	"context"
	"fmt"
	"log"

	searchdataindexscheduler "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/build/jobs/search_data_index_scheduler"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"

	_ "go.uber.org/automaxprocs"
)

func doTheThing(ctx context.Context) error {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.SearchDataIndexSchedulerConfig]()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	cfg.Database.RunMigrations = false

	scheduler, err := searchdataindexscheduler.Build(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error building scheduler: %w", err)
	}

	if err = scheduler.IndexTypes(ctx); err != nil {
		return fmt.Errorf("error indexing types: %w", err)
	}

	return nil
}

func main() {
	if err := doTheThing(context.Background()); err != nil {
		log.Fatal(err)
	}
}
