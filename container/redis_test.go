package container_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/worldline-go/test/container"
)

type RedisSuite struct {
	suite.Suite
	container *container.PostgresContainer
}

func (s *RedisSuite) SetupSuite() {
	s.container = container.Postgres(s.T())
}

func TestExampleTestSuiteRedis(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}

func (s *RedisSuite) TearDownSuite() {
	s.container.Stop(s.T())
}
