package container_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/worldline-go/test/container"
)

type KafkaSuite struct {
	suite.Suite
	container *container.KafkaContainer
}

func (s *KafkaSuite) SetupSuite() {
	s.container = container.Kafka(s.T())
}

func TestExampleTestSuiteKafka(t *testing.T) {
	suite.Run(t, new(KafkaSuite))
}

func (s *KafkaSuite) TearDownSuite() {
	s.container.Stop(s.T())
}

// func (s *KafkaSuite) TestTo() {
// 	s.T().Log("TestX")
// }
