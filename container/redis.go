package container

import (
	"net"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var DefaultRedisImage = "docker.dragonflydb.io/dragonflydb/dragonfly:v1.27.1"

type RedisContainer struct {
	container testcontainers.Container

	address []string
}

func (p *RedisContainer) Stop(t *testing.T) {
	t.Helper()

	if err := p.container.Terminate(t.Context()); err != nil {
		t.Fatalf("could not stop redis container: %v", err)
	}
}

func Redis(t *testing.T) *RedisContainer {
	t.Helper()

	image := DefaultPostgresImage
	if v := os.Getenv("TEST_IMAGE_REDIS"); v != "" {
		image = v
	}

	announceIP := "localhost"
	if v := os.Getenv("TESTCONTAINERS_HOST_OVERRIDE"); v != "" {
		announceIP = v
	}

	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: image,
			Cmd: []string{
				"dragonfly",
				"--logtostderr",
				"--cluster_mode=emulated",
				"--cluster_announce_ip=" + announceIP,
			},
			WaitingFor:   wait.ForLog("listening on port 6379"),
			ExposedPorts: []string{"6379/tcp"},
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.PortBindings = nat.PortMap{
					"6379/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: "6379",
						},
					},
				}
			},
		},
		Started:      true,
		ProviderType: 0,
		Reuse:        false,
	})
	if err != nil {
		t.Fatalf("could not create redis container: %v", err)
	}

	host, err := container.Host(t.Context())
	if err != nil {
		t.Fatalf("could not get host: %v", err)
	}

	address := net.JoinHostPort(host, "6379")

	return &RedisContainer{
		container: container,
		address:   []string{address},
	}
}

func (p *RedisContainer) Address() []string {
	return p.address
}
