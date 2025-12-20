package containerpostgres_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/worldline-go/test/container/containerpostgres"
)

type PostgresSuite struct {
	suite.Suite
	container *containerpostgres.Container
}

func (s *PostgresSuite) SetupSuite() {
	s.container = containerpostgres.New(s.T())
}

func TestExampleTestSuitePostgres(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}

func (s *PostgresSuite) TearDownSuite() {
	s.container.Stop(s.T())
}

func (s *PostgresSuite) SetupTest() {
	s.container.ExecuteFiles(s.T(), []string{"testdata/init.sql"})
}

func (s *PostgresSuite) TearDownTest() {
	sql := "Drop schema if exists transaction cascade;"
	_, err := s.container.Sql().Exec(sql)
	require.NoError(s.T(), err)
}
