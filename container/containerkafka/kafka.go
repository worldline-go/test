package containerkafka

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/worldline-go/test/utils"
	"github.com/worldline-go/test/utils/kafkautils"
	"github.com/worldline-go/wkafka"
)

var DefaultKafkaImage = "docker.io/bitnamilegacy/kafka:3.8.1"

type Container struct {
	container testcontainers.Container
	*kafkautils.KafkaTest

	address []string
}

func (p *Container) StopWithCtx(ctx context.Context) error {
	if p.KafkaTest != nil && p.KafkaTest.Client != nil {
		p.KafkaTest.Client.Close()
	}

	if p.container != nil {
		if err := p.container.Terminate(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Container) Stop(t *testing.T) {
	t.Helper()

	if err := p.StopWithCtx(t.Context()); err != nil {
		t.Fatalf("could not stop Kafka container: %v", err)
	}
}

func NewWithCtx(ctx context.Context) (*Container, error) {
	var kafkaContainer testcontainers.Container

	var addr []string
	if v := os.Getenv("KAFKA_BROKER"); v != "" {
		addr = strings.Fields(strings.ReplaceAll(v, ",", " "))
	}

	if len(addr) == 0 {
		image := DefaultKafkaImage
		if v := os.Getenv("TEST_IMAGE_KAFKA"); v != "" {
			image = v
		}

		announceIP := utils.DockerHost()

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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
					hostConfig.PortBindings = network.PortMap{
						network.MustParsePort("9092/tcp"): []network.PortBinding{
							{
								HostIP:   netip.MustParseAddr("0.0.0.0"),
								HostPort: "9092",
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
			return nil, fmt.Errorf("could not create Kafka container: %w", err)
		}

		host, err := container.Host(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not get host: %w", err)
		}

		addr = []string{net.JoinHostPort(host, "9092")}
		kafkaContainer = container
	}

	kafka, err := kafkautils.New(ctx, wkafka.Config{Brokers: addr})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &Container{
		container: kafkaContainer,
		address:   addr,
		KafkaTest: &kafkautils.KafkaTest{
			Kafka: kafka,
		},
	}, nil
}

func New(t *testing.T) *Container {
	t.Helper()

	kafkaContainer, err := NewWithCtx(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("kafka broker: %s", kafkaContainer.Address())

	return kafkaContainer
}

func (p *Container) Address() []string {
	return p.address
}
