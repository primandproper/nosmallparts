package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	tracingName = "db_client"
)

// Client is the primary database querying client.
type Client struct {
	tracer   tracing.Tracer
	logger   logging.Logger
	timeFunc func() time.Time
	config   database.ClientConfig
	db       *sql.DB
}

// ProvideDatabaseClient provides a new DataManager client.
func ProvideDatabaseClient(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider, cfg database.ClientConfig) (database.Client, error) {
	tracer := tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(tracingName))

	_, span := tracer.StartSpan(ctx)
	defer span.End()

	db, err := otelsql.Open("pgx", cfg.GetConnectionString(), otelsql.WithAttributes(
		attribute.KeyValue{
			Key:   semconv.ServiceNameKey,
			Value: attribute.StringValue("database"),
		},
	))
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres database: %w", err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(7)
	db.SetConnMaxLifetime(30 * time.Minute)

	c := &Client{
		db:       db,
		config:   cfg,
		tracer:   tracer,
		timeFunc: defaultTimeFunc,
		logger:   logging.EnsureLogger(logger).WithName("querier"),
	}

	return c, nil
}

// DB provides the database object.
func (q *Client) DB() *sql.DB {
	return q.db
}

// ReadDB provides the database object.
func (q *Client) ReadDB() *sql.DB {
	return q.db
}

// WriteDB provides the database object.
func (q *Client) WriteDB() *sql.DB {
	return q.db
}

// Close closes the database connection.
func (q *Client) Close() error {
	if err := q.db.Close(); err != nil {
		q.logger.Error("closing database connection", err)
		return err
	}

	return nil
}

// IsReady returns whether the database is ready for the querier.
func (q *Client) IsReady(ctx context.Context) bool {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("connection_url", q.config.GetConnectionString())

	attemptCount := 0
	for {
		if err := q.db.PingContext(ctx); err != nil {
			logger.WithValue("attempt_count", attemptCount).Info("ping failed, waiting for db")
			time.Sleep(q.config.GetPingWaitPeriod())

			attemptCount++
			if attemptCount >= int(q.config.GetMaxPingAttempts()) {
				break
			}
		} else {
			return true
		}
	}

	return false
}

func defaultTimeFunc() time.Time {
	return time.Now()
}

func (q *Client) CurrentTime() time.Time {
	if q == nil || q.timeFunc == nil {
		return defaultTimeFunc()
	}

	return q.timeFunc()
}

func (q *Client) RollbackTransaction(ctx context.Context, tx database.SQLQueryExecutorAndTransactionManager) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	q.logger.Debug("rolling back transaction")

	if err := tx.Rollback(); err != nil {
		observability.AcknowledgeError(err, q.logger, span, "rolling back transaction")
	}

	q.logger.Debug("transaction rolled back")
}
