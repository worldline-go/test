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

	"github.com/worldline-go/test/container/containerpostgres"
)

type DatabaseSuite struct {
	suite.Suite
	container *containerpostgres.Container
}

func (s *DatabaseSuite) SetupSuite() {
	s.container = containerpostgres.New(s.T())
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}

func (s *DatabaseSuite) TearDownSuite() {
	s.container.Stop(s.T())
}

func (s *DatabaseSuite) SetupTest() {
	s.container.ExecuteFiles(s.T(), []string{"testdata/init.sql"})
}

func (s *DatabaseSuite) TestAddEvents() {
	// Test event
}
```
