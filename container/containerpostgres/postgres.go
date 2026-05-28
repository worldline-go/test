package containerpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/worldline-go/test/utils"
	"github.com/worldline-go/test/utils/dbutils"
)

var DefaultPostgresImage = "docker.io/postgres:14.19-alpine"

type Container struct {
	container *postgres.PostgresContainer
	*dbutils.DatabaseTest

	address string
	dsn     string

	sql *sql.DB
}

func (p *Container) StopWithCtx(ctx context.Context) error {
	if p.sql != nil {
		if err := p.sql.Close(); err != nil {
			return fmt.Errorf("could not close sql connection: %w", err)
		}
	}

	if p.container != nil {
		if err := p.container.Terminate(ctx); err != nil {
			return fmt.Errorf("could not stop postgres container: %w", err)
		}
	}

	return nil
}

func (p *Container) Stop(t *testing.T) {
	t.Helper()

	if err := p.StopWithCtx(t.Context()); err != nil {
		t.Fatalf("could not stop postgres container: %v", err)
	}
}

func (p *Container) Sql() *sql.DB {
	return p.sql
}

func (p *Container) Address() string {
	return p.address
}

func (p *Container) DSN() string {
	return p.dsn
}

func NewWithCtx(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	image := DefaultPostgresImage
	if v := os.Getenv("TEST_IMAGE_POSTGRES"); v != "" {
		image = v
	}

	// Create options slice with defaults
	defaultOpts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("testdb"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		}),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
		testcontainers.WithLabels(utils.EnvToLabels()),
	}

	// Merge custom options with defaults
	allOpts := append(defaultOpts, opts...)

	// Run with merged options
	postgresContainer, err := postgres.Run(ctx, image, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("could not create Postgres container: %w", err)
	}

	// Get connection string
	addr, err := postgresContainer.PortEndpoint(ctx, "5432/tcp", "")
	if err != nil {
		return nil, fmt.Errorf("could not get Postgres port endpoint: %w", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get Postgres connection string: %w", err)
	}

	// Connect to database
	dbSql, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	if err := dbSql.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("could not ping to postgres: %w", err)
	}

	return &Container{
		container:    postgresContainer,
		address:      addr,
		dsn:          connStr,
		sql:          dbSql,
		DatabaseTest: &dbutils.DatabaseTest{DB: dbutils.New(dbSql)},
	}, nil
}

func New(t *testing.T, opts ...testcontainers.ContainerCustomizer) *Container {
	t.Helper()

	postgresContainer, err := NewWithCtx(t.Context(), opts...)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("postgres host: %s", postgresContainer.Address())
	t.Logf("postgres dsn: %s", postgresContainer.DSN())

	return postgresContainer
}

func (p *Container) CreateSnapshot(ctx context.Context) error {
	return p.container.Snapshot(ctx)
}

func (p *Container) RestoreSnapshot(ctx context.Context) error {
	return p.container.Restore(ctx)
}
