package container

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jmoiron/sqlx"
	"github.com/worldline-go/test/testdb"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	container testcontainers.Container
	*testdb.Database

	address string
	dsn     string

	sqlx *sqlx.DB
}

func (p *PostgresContainer) Stop(t *testing.T) {
	t.Helper()

	if err := p.container.Terminate(t.Context()); err != nil {
		t.Fatalf("could not stop postgres container: %v", err)
	}
}

func (p *PostgresContainer) Sqlx() *sqlx.DB {
	return p.sqlx
}

func (p *PostgresContainer) Address() string {
	return p.address
}

func (p *PostgresContainer) DSN() string {
	return p.dsn
}

func Postgres(t *testing.T) *PostgresContainer {
	addr := os.Getenv("POSTGRES_HOST")
	var postgresContainer testcontainers.Container

	if addr == "" {
		req := testcontainers.ContainerRequest{
			Image:        "postgres:13.15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_HOST_AUTH_METHOD": "trust",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
		}
		container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})

		if container == nil {
			t.Fatalf("could not create postgres container: %v", err)
		}

		port, err := container.MappedPort(t.Context(), "5432")
		if err != nil {
			t.Fatalf("could not get mapped port: %v", err)
		}

		host, err := container.Host(t.Context())
		if err != nil {
			t.Fatalf("could not get host: %v", err)
		}

		addr = net.JoinHostPort(host, port.Port())

		postgresContainer = container
	}

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s/postgres", addr)

	t.Logf("postgres host: %s", addr)
	t.Logf("postgres dsn: %s", dsn)

	dbSqlx, err := sqlx.ConnectContext(t.Context(), "pgx", dsn)
	if err != nil {
		t.Fatalf("could not connect to postgres: %v", err)
	}

	t.Logf("postgres connected at %s", dsn)

	return &PostgresContainer{
		container: postgresContainer,
		address:   addr,
		dsn:       dsn,
		sqlx:      dbSqlx,
		Database: &testdb.Database{
			DB: dbSqlx.DB,
		},
	}
}
