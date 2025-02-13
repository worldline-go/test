package container_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/worldline-go/test/container"
)

type PostgresSuite struct {
	suite.Suite
	container *container.PostgresContainer
}

func (s *PostgresSuite) SetupSuite() {
	s.container = container.Postgres(s.T())
}

func TestExampleTestSuite(t *testing.T) {
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
	_, err := s.container.Sqlx().Exec(sql)
	require.NoError(s.T(), err)
}
