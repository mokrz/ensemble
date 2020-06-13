package node

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
)

// Container wraps containerd.Container
type Container interface {
	ID() string
	Image(context.Context) (Image, error)
	Task(context.Context, cio.Attach) (Task, error)
}

func newContainer(c containerd.Container) Container {
	return &container{
		ctrContainer: c,
	}
}

type container struct {
	ctrContainer containerd.Container
}

func (c *container) ID() string {
	return c.ctrContainer.ID()
}

func (c *container) Image(ctx context.Context) (Image, error) {
	return c.ctrContainer.Image(ctx)
}

func (c *container) Task(ctx context.Context, attach cio.Attach) (Task, error) {
	task, err := c.ctrContainer.Task(ctx, attach)

	return newTask(task), err
}
