package node

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
)

// ExitStatus wraps containerd.ExitStatus
type ExitStatus containerd.ExitStatus

// Status wraps containerd.Status
type Status containerd.Status

// ProcessInfo wraps containerd.ProcessInfo
type ProcessInfo containerd.ProcessInfo

// Task wraps containerd.Task
type Task interface {
	ID() string
	Pid() uint32
	Status(ctx context.Context, attach cio.Attach) (Status, error)
	Pids(ctx context.Context) ([]ProcessInfo, error)
}

func newTask(c containerd.Task) Task {
	return &task{
		ctrTask: c,
	}
}

type task struct {
	ctrTask containerd.Task
}

func (t *task) ID() string {
	return t.ctrTask.ID()
}

func (t *task) Pid() uint32 {
	return t.ctrTask.Pid()
}

func (t *task) Status(ctx context.Context, attach cio.Attach) (Status, error) {
	stat, err := t.ctrTask.Status(ctx)
	return Status(stat), err
}

func (t *task) Pids(ctx context.Context) (pis []ProcessInfo, err error) {
	infos, err := t.ctrTask.Pids(ctx)

	for _, pi := range infos {
		pis = append(pis, ProcessInfo(pi))
	}

	return pis, err
}
