package containerkafka

import (
	"net"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var DefaultKafkaImage = "docker.io/bitnami/kafka:3.8.1"

type Container struct {
	container testcontainers.Container

	address []string
}

func (p *Container) Stop(t *testing.T) {
	t.Helper()

	if err := p.container.Terminate(t.Context()); err != nil {
		t.Fatalf("could not stop Kafka container: %v", err)
	}
}

func New(t *testing.T) *Container {
	t.Helper()

	image := DefaultKafkaImage
	if v := os.Getenv("TEST_IMAGE_KAFKA"); v != "" {
		image = v
	}

	announceIP := "localhost"
	if v := os.Getenv("TESTCONTAINERS_HOST_OVERRIDE"); v != "" {
		announceIP = v
	}

	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: image,
			Env: map[string]string{
				"ALLOW_PLAINTEXT_LISTENER":                 "yes",
				"KAFKA_CFG_NODE_ID":                        "0",
				"KAFKA_CFG_PROCESS_ROLES":                  "controller,broker",
				"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS":       "0@:9093",
				"KAFKA_CFG_LISTENERS":                      "PLAINTEXT://:9092,CONTROLLER://:9093,INTERNAL://:9094",
				"KAFKA_CFG_ADVERTISED_LISTENERS":           "PLAINTEXT://" + announceIP + ":9092,INTERNAL://kafka:9094",
				"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP": "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,INTERNAL:PLAINTEXT",
				"KAFKA_CFG_CONTROLLER_LISTENER_NAMES":      "CONTROLLER",
			},
			WaitingFor:   wait.ForLog("Kafka Server started"),
			ExposedPorts: []string{"9092/tcp"},
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.PortBindings = nat.PortMap{
					"9092/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: "9092",
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
		t.Fatalf("could not create Kafka container: %v", err)
	}

	host, err := container.Host(t.Context())
	if err != nil {
		t.Fatalf("could not get host: %v", err)
	}

	address := net.JoinHostPort(host, "9092")

	return &Container{
		container: container,
		address:   []string{address},
	}
}

func (p *Container) Address() []string {
	return p.address
}
