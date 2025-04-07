package containerredis_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/worldline-go/test/container/containerredis"
	container "github.com/worldline-go/test/container/containerredis"
)

type RedisSuite struct {
	suite.Suite
	container *containerredis.Container
}

func (s *RedisSuite) SetupSuite() {
	s.container = container.New(s.T())
}

func TestExampleTestSuiteRedis(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}

func (s *RedisSuite) TearDownSuite() {
	s.container.Stop(s.T())
}
