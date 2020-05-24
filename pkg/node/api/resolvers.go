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
	Name string `json:"name"`
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
	ID          string   `json:"id"`
	ContainerID string   `json:"container_id"`
	PID         uint32   `json:"pid"`
	PIDs        []uint32 `json:"pids"`
	Status      string   `json:"status"`
}

func getImageInfo(ctx context.Context, image containerd.Image) Image {
	return Image{
		Name: image.Name(),
	}
}

func newImageResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref           string
			image                    containerd.Image
			namespaceValid, refValid bool
			getImageErr              error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if ref, refValid = p.Args["ref"].(string); !refValid {
			return nil, errors.New("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if image, getImageErr = sp.GetImage(ctx, ref); getImageErr != nil {
			return nil, errors.New("image resolver failed with error: " + getImageErr.Error())
		}

		return getImageInfo(ctx, image), nil
	}
}

func newImagesResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter           string
			images                      []containerd.Image
			namespaceValid, filterValid bool
			getImagesErr                error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if p.Args["filter"] != nil {

			if filter, filterValid = p.Args["filter"].(string); !filterValid {
				return nil, errors.New("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if images, getImagesErr = sp.GetImages(ctx, filter); getImagesErr != nil {
			return nil, errors.New("images resolver failed with error: " + getImagesErr.Error())
		}

		var decoratedImages []Image
		for _, image := range images {
			decoratedImages = append(decoratedImages, getImageInfo(ctx, image))
		}

		return decoratedImages, nil
	}
}

func getContainerInfo(ctx context.Context, container containerd.Container) Container {
	var (
		containerTask           containerd.Task
		containerImage          containerd.Image
		getTaskErr, getImageErr error
	)
	if containerImage, getImageErr = container.Image(ctx); getImageErr != nil {
		return Container{}
	}
	if containerTask, getTaskErr = container.Task(ctx, nil); getTaskErr != nil {
		return Container{}
	}

	return Container{
		ID:    container.ID(),
		Image: getImageInfo(ctx, containerImage),
		Task:  getTaskInfo(ctx, containerTask),
	}
}

func newContainerResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, id           string
			container               containerd.Container
			namespaceValid, idValid bool
			getContainerErr         error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if id, idValid = p.Args["id"].(string); !idValid {
			return nil, errors.New("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if container, getContainerErr = sp.GetContainer(ctx, id); getContainerErr != nil {
			return nil, errors.New("container resolver failed with error: " + getContainerErr.Error())
		}

		return getContainerInfo(ctx, container), nil
	}
}

func newContainersResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter           string
			containers                  []containerd.Container
			namespaceValid, filterValid bool
			getContainersErr            error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if p.Args["filter"] != nil {

			if filter, filterValid = p.Args["filter"].(string); !filterValid {
				return nil, errors.New("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if containers, getContainersErr = sp.GetContainers(ctx, filter); getContainersErr != nil {
			return nil, errors.New("containers resolver failed with error: " + getContainersErr.Error())
		}

		var decoratedContainers []Container
		for _, container := range containers {
			decoratedContainers = append(decoratedContainers, getContainerInfo(ctx, container))
		}

		return decoratedContainers, nil
	}
}

func getPIDs(s []containerd.ProcessInfo, e error) (ret []uint32) {
	if e != nil {
		return
	}
	for _, i := range s {
		ret = append(ret, i.Pid)
	}
	return
}

func getTaskInfo(ctx context.Context, task containerd.Task) Task {
	return Task{
		ID:     task.ID(),
		Status: node.TaskStatus(ctx, task),
		PIDs:   getPIDs(task.Pids(ctx)),
		PID:    task.Pid(),
	}
}

func newTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			task                             containerd.Task
			namespaceValid, containerIDValid bool
			getTaskErr                       error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if p.Args["container_id"] != nil {

			if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
				return nil, errors.New("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if task, getTaskErr = sp.GetTask(ctx, containerID); getTaskErr != nil {
			return nil, errors.New("container resolver failed with error: " + getTaskErr.Error())
		}

		return getTaskInfo(ctx, task), nil
	}
}

func newTasksResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter           string
			tasks                       []containerd.Task
			filterValid, namespaceValid bool
			getTasksErr                 error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if p.Args["filter"] != nil {

			if filter, filterValid = p.Args["filter"].(string); !filterValid {
				return nil, errors.New("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if tasks, getTasksErr = sp.GetTasks(ctx, filter); getTasksErr != nil {
			return nil, errors.New("containers resolver failed with error: " + getTasksErr.Error())
		}

		var decoratedTasks []Task
		for _, task := range tasks {
			decoratedTasks = append(decoratedTasks, getTaskInfo(ctx, task))
		}

		return decoratedTasks, nil
	}
}

func newCreateImageResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref           string
			image                    containerd.Image
			namespaceValid, refValid bool
			imagePullErr             error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if ref, refValid = p.Args["ref"].(string); !refValid {
			return nil, errors.New("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if image, imagePullErr = sp.CreateImage(ctx, ref); imagePullErr != nil {
			return nil, errors.New("createImage resolver failed to create image: " + ref + " with error:  " + imagePullErr.Error())
		}

		return getImageInfo(ctx, image), nil
	}
}

func newCreateContainerResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ID, imgRef             string
			container                         containerd.Container
			namespaceValid, refValid, IDValid bool
			containerCreateErr                error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if imgRef, refValid = p.Args["image_ref"].(string); !refValid {
			return nil, errors.New("invalid request")
		}

		if ID, IDValid = p.Args["id"].(string); !IDValid {
			return nil, errors.New("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if container, containerCreateErr = sp.CreateContainer(ctx, imgRef, ID); containerCreateErr != nil {
			return nil, errors.New("createContainer resolver failed to create container: " + ID + " with error:  " + containerCreateErr.Error())
		}

		return getContainerInfo(ctx, container), nil
	}
}

func newCreateTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			namespaceValid, containerIDValid bool
			task                             containerd.Task
			createTaskErr                    error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
			return nil, errors.New("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if task, createTaskErr = sp.CreateTask(ctx, containerID); createTaskErr != nil {
			return nil, errors.New("createTask resolver failed to create task for container: " + containerID + " with error:  " + createTaskErr.Error())
		}

		return getTaskInfo(ctx, task), nil
	}
}

func newKillTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			namespaceValid, containerIDValid bool
			killTaskErr                      error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
			return nil, errors.New("invalid request")
		}

		if killTaskErr = sp.KillTask(namespaces.WithNamespace(context.Background(), namespace), containerID); killTaskErr != nil {
			return nil, errors.New("killTask resolver failed with error: " + killTaskErr.Error())
		}

		return nil, nil
	}
}

func newDeleteImageResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref           string
			namespaceValid, refValid bool
			deleteImageErr           error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if ref, refValid = p.Args["ref"].(string); !refValid {
			return nil, errors.New("invalid request")
		}

		if deleteImageErr = sp.DeleteImage(namespaces.WithNamespace(context.Background(), namespace), ref); deleteImageErr != nil {
			return nil, errors.New("deleteImage resolver failed with error: " + deleteImageErr.Error())
		}

		return nil, nil
	}
}

func newDeleteContainerResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ID           string
			namespaceValid, IDValid bool
			deleteContainerErr      error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if ID, IDValid = p.Args["id"].(string); !IDValid {
			return nil, errors.New("invalid request")
		}

		if deleteContainerErr = sp.DeleteContainer(namespaces.WithNamespace(context.Background(), namespace), ID); deleteContainerErr != nil {
			return nil, errors.New("deleteContainer resolver failed with error: " + deleteContainerErr.Error())
		}

		return nil, nil
	}
}

func newDeleteTaskResolver(sp node.Service) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			exitStatus                       *containerd.ExitStatus
			namespaceValid, containerIDValid bool
			deleteTaskErr                    error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, errors.New("invalid request")
		}

		if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
			return nil, errors.New("invalid request")
		}

		if exitStatus, deleteTaskErr = sp.DeleteTask(namespaces.WithNamespace(context.Background(), namespace), containerID); deleteTaskErr != nil {
			return nil, errors.New("deleteTask resolver failed with error: " + deleteTaskErr.Error())
		}

		return exitStatus, nil
	}
}
