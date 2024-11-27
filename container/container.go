package container

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

type container struct {
	Container testcontainers.Container
}

func (c *container) Close() {
	if c.Container != nil {
		c.Container.Terminate(context.Background())
	}
}
