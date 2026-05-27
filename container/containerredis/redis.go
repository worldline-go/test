package containerredis

import (
	"net"
	"net/netip"
	"os"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/worldline-go/test/utils"
)

var DefaultRedisImage = "docker.dragonflydb.io/dragonflydb/dragonfly:v1.27.1"

type Container struct {
	container testcontainers.Container

	address []string
}

func (p *Container) Stop(t *testing.T) {
	t.Helper()

	if err := p.container.Terminate(t.Context()); err != nil {
		t.Fatalf("could not stop redis container: %v", err)
	}
}

func New(t *testing.T) *Container {
	t.Helper()

	image := DefaultRedisImage
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
			hostConfig.PortBindings = network.PortMap{
				network.MustParsePort("6379/tcp"): []network.PortBinding{
					{
						HostIP:   netip.MustParseAddr("0.0.0.0"),
						HostPort: "6379",
					},
				},
			}
			},
			Labels: utils.EnvToLabels(),
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

	return &Container{
		container: container,
		address:   []string{address},
	}
}

func (p *Container) Address() []string {
	return p.address
}
