package containerkafka_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	container "github.com/worldline-go/test/container/containerkafka"
)

type KafkaSuite struct {
	suite.Suite
	container *container.Container
}

func (s *KafkaSuite) SetupSuite() {
	s.container = container.New(s.T())
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
