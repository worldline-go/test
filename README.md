# test 🧪

Testing with using real servers and databases with test containers.

```sh
go get github.com/worldline-go/test
```

| Env                 | Default Image                                       |
| ------------------- | --------------------------------------------------- |
| TEST_IMAGE_POSTGRES | docker.io/postgres:13.15-alpine                     |
| TEST_IMAGE_REDIS    | docker.dragonflydb.io/dragonflydb/dragonfly:v1.27.1 |
| TEST_IMAGE_KAFKA    | docker.io/bitnami/kafka:3.8.1					    |

## PostgreSQL

Need to have a running PostgreSQL database to run the tests. To do that run it in the package level test main function.

Our test package has `Main` function it will run the given function and accept a defer function for cleanup.

```go
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
```
