package log

import (
	"context"
	"time"
	"errors"

	"github.com/containerd/containerd/namespaces"
	"github.com/mokrz/clamor/node"
	"go.uber.org/zap"
)

// NewLoggingNode wraps the given node.Service in logging middleware
func NewLoggingNode(logger *zap.Logger, svc node.Service) node.Service {
	return &loggingNode{
		logger: logger,
		next:   svc,
	}
}

type loggingNode struct {
	logger *zap.Logger
	next   node.Service
}

func baseFields(ctx context.Context) []zap.Field {
	var (
		logFields []zap.Field
		ns        string
		nsExists  bool
	)

	if ns, nsExists = namespaces.Namespace(ctx); nsExists {
		logFields = append(logFields, zap.String("namespace", ns))
	}

	return logFields
}

func (ln *loggingNode) PullImage(ctx context.Context, name string) (image node.Image, err error) {
	var dne node.ErrNotFound
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("image", name))
	msg := "PullImage"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if image, err = ln.next.PullImage(ctx, name); err == nil {
			ln.logger.Info(msg, logFields...)
		} else if errors.As(err, &dne) {
			logFields = append(logFields, zap.String("error", dne.Error()))
			ln.logger.Warn(msg, logFields...)
		} else {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		}

	}(time.Now())

	return image, err
}

func (ln *loggingNode) GetImage(ctx context.Context, name string) (image node.Image, err error) {
	var dne node.ErrNotFound
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("image", name))
	msg := "GetImage"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if image, err = ln.next.GetImage(ctx, name); err == nil {
			ln.logger.Info(msg, logFields...)
		} else if errors.As(err, &dne) {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Warn(msg, logFields...)
		} else {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		}

	}(time.Now())

	return image, err
}

func (ln *loggingNode) GetImages(ctx context.Context, filter string) (images []node.Image, err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("filter", filter))
	msg := "GetImages"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if images, err = ln.next.GetImages(ctx, filter); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}

	}(time.Now())

	return images, err
}

func (ln *loggingNode) DeleteImage(ctx context.Context, name string) (err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("image", name))
	msg := "DeleteImage"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if err = ln.next.DeleteImage(ctx, name); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return err
}

func (ln *loggingNode) CreateContainer(ctx context.Context, imageName string, id string) (container node.Container, err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("image", imageName))
	logFields = append(logFields, zap.String("id", id))
	msg := "CreateContainer"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if container, err = ln.next.CreateContainer(ctx, imageName, id); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return container, err
}

func (ln *loggingNode) GetContainer(ctx context.Context, id string) (container node.Container, err error) {
	var dne node.ErrNotFound
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("id", id))
	msg := "GetContainer"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if container, err = ln.next.GetContainer(ctx, id); err == nil {
			ln.logger.Info(msg, logFields...)
		} else if errors.As(err, &dne) {
			logFields = append(logFields, zap.String("error", dne.Error()))
			ln.logger.Warn(msg, logFields...)
		} else {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		}
		
	}(time.Now())

	return container, err
}

func (ln *loggingNode) GetContainers(ctx context.Context, filter string) (containers []node.Container, err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("filter", filter))
	msg := "GetContainers"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if containers, err = ln.next.GetContainers(ctx, filter); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return containers, err
}

func (ln *loggingNode) DeleteContainer(ctx context.Context, id string) (err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("id", id))
	msg := "DeleteContainer"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if err = ln.next.DeleteContainer(ctx, id); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return err
}

func (ln *loggingNode) CreateTask(ctx context.Context, containerID string) (task node.Task, err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("container_id", containerID))
	msg := "CreateTask"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if task, err = ln.next.CreateTask(ctx, containerID); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return task, err
}

func (ln *loggingNode) GetTask(ctx context.Context, containerID string) (task node.Task, err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("container_id", containerID))
	msg := "GetTask"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if task, err = ln.next.GetTask(ctx, containerID); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return task, err
}

func (ln *loggingNode) GetTasks(ctx context.Context, filter string) (tasks []node.Task, err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("filter", filter))
	msg := "GetTasks"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if tasks, err = ln.next.GetTasks(ctx, filter); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return tasks, err
}

func (ln *loggingNode) KillTask(ctx context.Context, containerID string) (err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("container_id", containerID))
	msg := "KillTask"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if err = ln.next.KillTask(ctx, containerID); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return err
}

func (ln *loggingNode) DeleteTask(ctx context.Context, containerID string) (exitStatus node.ExitStatus, err error) {
	logFields := baseFields(ctx)
	logFields = append(logFields, zap.String("container_id", containerID))
	msg := "DeleteTask"

	defer func(took time.Time) {
		logFields = append(logFields, zap.String("took", time.Since(took).String()))

		if exitStatus, err = ln.next.DeleteTask(ctx, containerID); err != nil {
			logFields = append(logFields, zap.String("error", err.Error()))
			ln.logger.Error(msg, logFields...)
		} else {
			ln.logger.Info(msg, logFields...)
		}
	}(time.Now())

	return exitStatus, err
}
