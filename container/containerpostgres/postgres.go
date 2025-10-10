package containerpostgres

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/worldline-go/test/utils/dbutils"
)

var DefaultPostgresImage = "docker.io/postgres:14.19-alpine"

type Container struct {
	container *postgres.PostgresContainer
	*dbutils.DatabaseTest

	address string
	dsn     string

	sqlx *sqlx.DB
}

func (p *Container) Stop(t *testing.T) {
	t.Helper()

	if p.sqlx != nil {
		if err := p.sqlx.Close(); err != nil {
			t.Errorf("could not close sqlx connection: %v", err)
		}
	}

	if p.container != nil {
		if err := p.container.Terminate(t.Context()); err != nil {
			t.Fatalf("could not stop postgres container: %v", err)
		}
	}
}

func (p *Container) Sqlx() *sqlx.DB {
	return p.sqlx
}

func (p *Container) Address() string {
	return p.address
}

func (p *Container) DSN() string {
	return p.dsn
}

func New(t *testing.T, port string) *Container {
	t.Helper()

	//addr := os.Getenv("POSTGRES_HOST")

	image := DefaultPostgresImage
	if v := os.Getenv("TEST_IMAGE_POSTGRES"); v != "" {
		image = v
	}

	if port == "" {
		port = "5432/tcp"
	}

	// Create Postgres container
	postgresContainer, err := postgres.Run(t.Context(),
		image,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(time.Second*5)),
		testcontainers.WithExposedPorts(port),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(t.Context(), "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	// Connect to database
	dbSqlx, err := sqlx.ConnectContext(t.Context(), "pgx", connStr)
	if err != nil {
		t.Fatalf("could not connect to postgres: %v", err)
	}

	// Test connection
	if err := dbSqlx.Ping(); err != nil {
		t.Fatal(err)
	}

	t.Logf("postgres connection string: %s", connStr)

	return &Container{
		container:    postgresContainer,
		address:      connStr,
		sqlx:         dbSqlx,
		DatabaseTest: dbutils.NewTest(t, dbSqlx.DB),
	}
}

func (p *Container) CreateSnapshot(ctx context.Context) error {
	return p.container.Snapshot(ctx)
}

func (p *Container) RestoreSnapshot(ctx context.Context) error {
	return p.container.Restore(ctx)
}
