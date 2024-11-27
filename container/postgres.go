package container

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type postgres struct {
	container

	Address string
	DSN     string
}

func Postgres(ctx context.Context) (*postgres, error) {
	addr := os.Getenv("POSTGRES_HOST")
	var postgresContainer testcontainers.Container

	if addr != "" {
		req := testcontainers.ContainerRequest{
			Image:        "postgres:13.15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_HOST_AUTH_METHOD": "trust",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Second),
		}
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})

		if container == nil {
			return nil, fmt.Errorf("could not create postgres container: %w", err)
		}

		port, err := container.MappedPort(ctx, "5432")
		if err != nil {
			return nil, err
		}

		host, err := container.Host(ctx)
		if err != nil {
			return nil, err
		}

		addr = net.JoinHostPort(host, port.Port())

		postgresContainer = container
	}

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s/postgres", addr)

	return &postgres{
		container: container{
			Container: postgresContainer,
		},
		Address: addr,
		DSN:     dsn,
	}, nil
}
