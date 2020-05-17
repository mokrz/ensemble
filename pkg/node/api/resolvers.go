package api

import (
	"context"
	"errors"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/pkg/node"
)

// Image instances hold metadata for a container image.
// TODO: Add image properties (size, age, etc.).
type Image struct {
	Name string `json:"ref"`
}

// Container instances hold metadata for a container.
// TODO: Add container properties (size, age, etc.).
type Container struct {
	ID    string `json:"id"`
	Image Image  `json:"image"`
	Task  Task   `json:"task"`
}

// Task instances hold metadata for a container task.
// TODO: Add task properties (PIDs, metrics, etc.).
type Task struct {
	ID          string `json:"id"`
	ContainerID string `json:"container_id"`
	Status      string `json:"status"`
}

func newImageResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref = p.Args["namespace"].(string), p.Args["ref"].(string)
			ctx            = namespaces.WithNamespace(context.Background(), namespace)
			image          containerd.Image
			getImageErr    error
		)

		if image, getImageErr = sp.GetImage(ctx, ref); getImageErr != nil {
			return nil, errors.New("image resolver failed with error: " + getImageErr.Error())
		}

		return &Image{
			Name: image.Name(),
		}, nil
	}
}

func newImagesResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter = p.Args["namespace"].(string), p.Args["filter"].(string)
			ctx               = namespaces.WithNamespace(context.Background(), namespace)
			images            []containerd.Image
			getImagesErr      error
		)

		if images, getImagesErr = sp.GetImages(ctx, filter); getImagesErr != nil {
			return nil, errors.New("images resolver failed with error: " + getImagesErr.Error())
		}

		var decoratedImages []Image
		for _, image := range images {
			decoratedImages = append(decoratedImages, Image{
				Name: image.Name(),
			})
		}

		return decoratedImages, nil
	}
}

func newContainerResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, id                     = p.Args["namespace"].(string), p.Args["id"].(string)
			ctx                               = namespaces.WithNamespace(context.Background(), namespace)
			container                         containerd.Container
			containerTask                     containerd.Task
			getContainerErr, containerTaskErr error
		)

		if container, getContainerErr = sp.GetContainer(ctx, id); getContainerErr != nil {
			return nil, errors.New("container resolver failed with error: " + getContainerErr.Error())
		}

		if containerTask, containerTaskErr = sp.GetTask(ctx, container.ID()); containerTaskErr != nil {
			return nil, errors.New("container resolver failed with error: " + containerTaskErr.Error())
		}

		return &Container{
			ID: container.ID(),
			Task: Task{
				ID:          containerTask.ID(),
				ContainerID: container.ID(),
				Status:      node.TaskStatus(ctx, containerTask),
			},
		}, nil
	}
}

func newContainersResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter = p.Args["namespace"].(string), p.Args["filter"].(string)
			ctx               = namespaces.WithNamespace(context.Background(), namespace)
			containers        []containerd.Container
			getContainersErr  error
		)

		if containers, getContainersErr = sp.GetContainers(ctx, filter); getContainersErr != nil {
			return nil, errors.New("containers resolver failed with error: " + getContainersErr.Error())
		}

		var decoratedContainers []Container
		for _, container := range containers {
			image, imageErr := container.Image(ctx)

			if imageErr != nil {
				return nil, errors.New("containers resolver failed to get image for container" + container.ID() + " with error: " + imageErr.Error())
			}

			decoratedContainers = append(decoratedContainers, Container{
				ID: container.ID(),
				Image: Image{
					Name: image.Name(),
				},
			})
		}

		return decoratedContainers, nil
	}
}

func newTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, id = p.Args["namespace"].(string), p.Args["id"].(string)
			ctx           = namespaces.WithNamespace(context.Background(), namespace)
			task          containerd.Task
			getTaskErr    error
		)

		if task, getTaskErr = sp.GetTask(ctx, id); getTaskErr != nil {
			return nil, errors.New("container resolver failed with error: " + getTaskErr.Error())
		}

		return &Task{
			ID:     task.ID(),
			Status: node.TaskStatus(ctx, task),
		}, nil
	}
}

func newTasksResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter = p.Args["namespace"].(string), p.Args["filter"].(string)
			ctx               = namespaces.WithNamespace(context.Background(), namespace)
			tasks             []containerd.Task
			getTasksErr       error
		)

		if tasks, getTasksErr = sp.GetTasks(ctx, filter); getTasksErr != nil {
			return nil, errors.New("containers resolver failed with error: " + getTasksErr.Error())
		}

		var decoratedTasks []Task
		for _, task := range tasks {
			stat, statErr := task.Status(ctx)

			if statErr != nil {
				return nil, errors.New("tasks resolver failed to get status for task" + task.ID() + " with error: " + statErr.Error())
			}

			decoratedTasks = append(decoratedTasks, Task{
				ID:     task.ID(),
				Status: string(stat.Status),
			})
		}

		return decoratedTasks, nil
	}
}

func newCreateImageResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref = p.Args["namespace"].(string), p.Args["ref"].(string)
			ctx            = namespaces.WithNamespace(context.Background(), namespace)
			image          containerd.Image
			imagePullErr   error
		)

		if image, imagePullErr = sp.CreateImage(ctx, ref); imagePullErr != nil {
			return nil, errors.New("createImage resolver failed to create image: " + ref + " with error:  " + imagePullErr.Error())
		}

		return Image{
			Name: image.Name(),
		}, nil
	}
}

func newCreateContainerResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, id, imgRef            = p.Args["namespace"].(string), p.Args["id"].(string), p.Args["image_ref"].(string)
			ctx                              = namespaces.WithNamespace(context.Background(), namespace)
			container                        containerd.Container
			image                            containerd.Image
			containerCreateErr, imagePullErr error
		)

		if image, imagePullErr = sp.CreateImage(ctx, imgRef); imagePullErr != nil {
			return nil, errors.New("createContainer resolver failed to pull image: " + imgRef + " with error:  " + imagePullErr.Error())
		}

		if container, containerCreateErr = sp.CreateContainer(ctx, imgRef, id); containerCreateErr != nil {
			return nil, errors.New("createContainer resolver failed to create container: " + id + " with error:  " + containerCreateErr.Error())
		}

		return Container{
			ID: container.ID(),
			Image: Image{
				Name: image.Name(),
			},
		}, nil
	}
}

func newCreateTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			ctx                            = namespaces.WithNamespace(context.Background(), p.Args["namespace"].(string))
			containerID                    = p.Args["container_id"].(string)
			container                      containerd.Container
			task                           containerd.Task
			createTaskErr, getContainerErr error
		)

		if container, getContainerErr = sp.GetContainer(ctx, containerID); getContainerErr != nil {
			return nil, errors.New("createTask resolver failed to get container: " + containerID + " with error:  " + getContainerErr.Error())
		}

		if task, createTaskErr = sp.CreateTask(ctx, containerID); createTaskErr != nil {
			return nil, errors.New("createTask resolver failed to create task for container: " + containerID + " with error:  " + createTaskErr.Error())
		}

		return Task{
			ID:          task.ID(),
			ContainerID: container.ID(),
			Status:      node.TaskStatus(ctx, task),
		}, nil
	}
}

func newKillTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID = p.Args["namespace"].(string), p.Args["container_id"].(string)
			ctx                    = namespaces.WithNamespace(context.Background(), namespace)
			killTaskErr            error
		)

		if killTaskErr = sp.KillTask(ctx, containerID); killTaskErr != nil {
			return nil, errors.New("killTask resolver failed with error: " + killTaskErr.Error())
		}

		return nil, nil
	}
}

func newDeleteImageResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref = p.Args["namespace"].(string), p.Args["ref"].(string)
			ctx            = namespaces.WithNamespace(context.Background(), namespace)
			deleteImageErr error
		)

		if deleteImageErr = sp.DeleteImage(ctx, ref); deleteImageErr != nil {
			return nil, errors.New("deleteImage resolver failed with error: " + deleteImageErr.Error())
		}

		return nil, nil
	}
}

func newDeleteContainerResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, id      = p.Args["namespace"].(string), p.Args["id"].(string)
			ctx                = namespaces.WithNamespace(context.Background(), namespace)
			deleteContainerErr error
		)

		if deleteContainerErr = sp.DeleteContainer(ctx, id); deleteContainerErr != nil {
			return nil, errors.New("deleteContainer resolver failed with error: " + deleteContainerErr.Error())
		}

		return nil, nil
	}
}

func newDeleteTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID = p.Args["namespace"].(string), p.Args["container_id"].(string)
			ctx                    = namespaces.WithNamespace(context.Background(), namespace)
			exitStatus             *containerd.ExitStatus
			deleteTaskErr          error
		)

		if exitStatus, deleteTaskErr = sp.DeleteTask(ctx, containerID); deleteTaskErr != nil {
			return nil, errors.New("deleteTask resolver failed with error: " + deleteTaskErr.Error())
		}

		return exitStatus, nil
	}
}
