package api

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/namespaces"
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/node"
)

// ResolverSet groups all the graphql.FieldResolveFn instances together.
type ResolverSet struct {
	CreateImageResolver,
	ImageResolver,
	ImagesResolver,
	DeleteImageResolver,
	CreateContainerResolver,
	ContainerResolver,
	ContainersResolver,
	DeleteContainerResolver,
	CreateTaskResolver,
	TaskResolver,
	TasksResolver,
	DeleteTaskResolver,
	KillTaskResolver graphql.FieldResolveFn
}

// NewResolverSet creates ResolverSet methods. The created resolvers interact with the node via the given node.Service implementation.
func NewResolverSet(svc node.Service) *ResolverSet {
	return &ResolverSet{
		CreateImageResolver:     NewCreateImageResolver(svc),
		ImageResolver:           NewImageResolver(svc),
		ImagesResolver:          NewImagesResolver(svc),
		DeleteImageResolver:     NewDeleteImageResolver(svc),
		CreateContainerResolver: NewCreateContainerResolver(svc),
		ContainerResolver:       NewContainerResolver(svc),
		ContainersResolver:      NewContainersResolver(svc),
		DeleteContainerResolver: NewDeleteContainerResolver(svc),
		CreateTaskResolver:      NewCreateTaskResolver(svc),
		TaskResolver:            NewTaskResolver(svc),
		TasksResolver:           NewTasksResolver(svc),
		DeleteTaskResolver:      NewDeleteTaskResolver(svc),
		KillTaskResolver:        NewKillTaskResolver(svc),
	}
}

// Image holds metadata for a container image.
// TODO: Add image properties (size, age, etc.).
type Image struct {
	Name string `json:"name"`
}

// Container holds metadata for a container.
// TODO: Add container properties (size, age, etc.).
type Container struct {
	ID    string `json:"id"`
	Image Image  `json:"image"`
	Task  Task   `json:"task"`
}

// Task holds metadata for a container task.
// TODO: Add task properties (PIDs, metrics, etc.).
type Task struct {
	ID          string   `json:"id"`
	ContainerID string   `json:"container_id"`
	PID         uint32   `json:"pid"`
	PIDs        []uint32 `json:"pids"`
	Status      string   `json:"status"`
}

func getImageInfo(ctx context.Context, i node.Image) Image {
	return Image{
		Name: i.Name(),
	}
}

// NewImageResolver returns a graphql resolver that looks up the given image name in the given namespace
func NewImageResolver(svc node.ImageService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref           string
			image                    node.Image
			namespaceValid, refValid bool
			getImageErr              error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, fmt.Errorf("invalid request")
		}

		if ref, refValid = p.Args["ref"].(string); !refValid {
			return nil, fmt.Errorf("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if image, getImageErr = svc.GetImage(ctx, ref); getImageErr != nil {
			return nil, fmt.Errorf("image resolver failed: %w", getImageErr)
		}

		return getImageInfo(ctx, image), nil
	}
}

// NewImagesResolver returns a graphql resolver that looks up all images in the given namespace
func NewImagesResolver(svc node.ImageService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter           string
			images                      []node.Image
			namespaceValid, filterValid bool
			getImagesErr                error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, fmt.Errorf("invalid request")
		}

		if p.Args["filter"] != nil {

			if filter, filterValid = p.Args["filter"].(string); !filterValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if images, getImagesErr = svc.GetImages(ctx, filter); getImagesErr != nil {
			return nil, fmt.Errorf("images resolver failed: %w", getImagesErr)
		}

		var decoratedImages []Image
		for _, image := range images {
			decoratedImages = append(decoratedImages, getImageInfo(ctx, image))
		}

		return decoratedImages, nil
	}
}

func getContainerInfo(ctx context.Context, c node.Container) Container {
	var (
		containerTask           node.Task
		containerImage          node.Image
		getTaskErr, getImageErr error
	)

	if containerImage, getImageErr = c.Image(ctx); getImageErr != nil {
		return Container{}
	}

	if containerTask, getTaskErr = c.Task(ctx, nil); getTaskErr != nil {
		return Container{
			ID:    c.ID(),
			Image: getImageInfo(ctx, containerImage),
			Task:  Task{},
		}
	}

	return Container{
		ID:    c.ID(),
		Image: getImageInfo(ctx, containerImage),
		Task:  getTaskInfo(ctx, containerTask),
	}
}

// NewContainerResolver returns a graphql resolver that looks up the given container ID in the given namespace
func NewContainerResolver(svc node.ContainerService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, id           string
			container               node.Container
			namespaceValid, idValid bool
			getContainerErr         error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, fmt.Errorf("invalid request")
		}

		if id, idValid = p.Args["id"].(string); !idValid {
			return nil, fmt.Errorf("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if container, getContainerErr = svc.GetContainer(ctx, id); getContainerErr != nil {
			return nil, fmt.Errorf("container resolver failed: %w", getContainerErr)
		}

		return getContainerInfo(ctx, container), nil
	}
}

// NewContainersResolver returns a graphql resolver that looks up all containers in the given namespace
func NewContainersResolver(svc node.ContainerService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter           string
			containers                  []node.Container
			namespaceValid, filterValid bool
			getContainersErr            error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, fmt.Errorf("invalid request")
		}

		if p.Args["filter"] != nil {

			if filter, filterValid = p.Args["filter"].(string); !filterValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if containers, getContainersErr = svc.GetContainers(ctx, filter); getContainersErr != nil {
			return nil, fmt.Errorf("containers resolver failed: %w", getContainersErr)
		}

		var decoratedContainers []Container
		for _, container := range containers {
			decoratedContainers = append(decoratedContainers, getContainerInfo(ctx, container))
		}

		return decoratedContainers, nil
	}
}

func getPIDs(s []node.ProcessInfo, e error) (ret []uint32) {
	if e != nil {
		return
	}
	for _, i := range s {
		ret = append(ret, i.Pid)
	}
	return
}

func getTaskInfo(ctx context.Context, t node.Task) Task {
	return Task{
		ID:     t.ID(),
		Status: node.TaskStatus(ctx, t),
		PIDs:   getPIDs(t.Pids(ctx)),
		PID:    t.Pid(),
	}
}

// NewTaskResolver returns a graphql resolver that looks up the current task for the given container ID in the given namespace
func NewTaskResolver(svc node.TaskService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			task                             node.Task
			namespaceValid, containerIDValid bool
			getTaskErr                       error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, fmt.Errorf("invalid request")
		}

		if p.Args["container_id"] != nil {

			if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if task, getTaskErr = svc.GetTask(ctx, containerID); getTaskErr != nil {
			return nil, fmt.Errorf("container resolver failed: %w", getTaskErr)
		}

		return getTaskInfo(ctx, task), nil
	}
}

// NewTasksResolver returns a graphql resolver that looks up all container tasks in the given namespace
func NewTasksResolver(svc node.TaskService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, filter           string
			tasks                       []node.Task
			filterValid, namespaceValid bool
			getTasksErr                 error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, fmt.Errorf("invalid request")
		}

		if p.Args["filter"] != nil {

			if filter, filterValid = p.Args["filter"].(string); !filterValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if tasks, getTasksErr = svc.GetTasks(ctx, filter); getTasksErr != nil {
			return nil, fmt.Errorf("containers resolver failed: %w", getTasksErr)
		}

		var decoratedTasks []Task
		for _, task := range tasks {
			decoratedTasks = append(decoratedTasks, getTaskInfo(ctx, task))
		}

		return decoratedTasks, nil
	}
}

// NewCreateImageResolver returns a graphql resolver that creates a container image from the given ref, pulling from remote registries if necessary
func NewCreateImageResolver(svc node.ImageService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref           string
			image                    node.Image
			namespaceValid, refValid bool
			imagePullErr             error
		)

		if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
			return nil, fmt.Errorf("invalid request")
		}

		if ref, refValid = p.Args["ref"].(string); !refValid {
			return nil, fmt.Errorf("invalid request")
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if image, imagePullErr = svc.PullImage(ctx, ref); imagePullErr != nil {
			return nil, fmt.Errorf("createImage resolver failed to create image %s: %w", ref, imagePullErr)
		}

		return getImageInfo(ctx, image), nil
	}
}

// NewCreateContainerResolver returns a graphql resolver that creates a container with the given ID from the given image
func NewCreateContainerResolver(sp node.ContainerService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ID, imageName                string
			container                               node.Container
			namespaceValid, imageNameValid, IDValid bool
			containerCreateErr                      error
		)

		if p.Args["namespace"] != nil {

			if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if p.Args["image"] != nil {

			if imageName, imageNameValid = p.Args["image"].(string); !imageNameValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if p.Args["id"] != nil {

			if ID, IDValid = p.Args["id"].(string); !IDValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if container, containerCreateErr = sp.CreateContainer(ctx, imageName, ID); containerCreateErr != nil {
			return nil, fmt.Errorf("createContainer resolver failed to create container %s: %w", ID, containerCreateErr)
		}

		return getContainerInfo(ctx, container), nil
	}
}

// NewCreateTaskResolver returns a graphql resolver that creates a task with the given ID from the given image
func NewCreateTaskResolver(ns node.TaskService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			namespaceValid, containerIDValid bool
			task                             node.Task
			createTaskErr                    error
		)

		if p.Args["namespace"] != nil {

			if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if p.Args["container_id"] != nil {

			if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		ctx := namespaces.WithNamespace(context.Background(), namespace)

		if task, createTaskErr = ns.CreateTask(ctx, containerID); createTaskErr != nil {
			return nil, fmt.Errorf("createTask resolver failed to create task for container %s: %w", containerID, createTaskErr)
		}

		return getTaskInfo(ctx, task), nil
	}
}

// NewKillTaskResolver returns a graphql resolver that kills the task associated with the given container
func NewKillTaskResolver(ns node.TaskService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			namespaceValid, containerIDValid bool
			killTaskErr                      error
		)

		if p.Args["namespace"] != nil {

			if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if p.Args["container_id"] != nil {

			if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if killTaskErr = ns.KillTask(namespaces.WithNamespace(context.Background(), namespace), containerID); killTaskErr != nil {
			return nil, fmt.Errorf("killTask resolver failed to kill task for %s: %w", containerID, killTaskErr)
		}

		return nil, nil
	}
}

// NewDeleteImageResolver returns a graphql resolver that deletes the given image
func NewDeleteImageResolver(ns node.ImageService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ref           string
			namespaceValid, refValid bool
			deleteImageErr           error
		)

		if p.Args["namespace"] != nil {

			if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
				return nil,fmt.Errorf("invalid request")
			}
		}

		if p.Args["ref"] != nil {

			if ref, refValid = p.Args["ref"].(string); !refValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if deleteImageErr = ns.DeleteImage(namespaces.WithNamespace(context.Background(), namespace), ref); deleteImageErr != nil {
			return nil, fmt.Errorf("deleteImage resolver failed to delete %s: %w", ref, deleteImageErr)
		}

		return nil, nil
	}
}

// NewDeleteContainerResolver returns a graphql resolver that deletes the given container
func NewDeleteContainerResolver(ns node.ContainerService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, ID           string
			namespaceValid, IDValid bool
			deleteContainerErr      error
		)

		if p.Args["namespace"] != nil {

			if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if p.Args["id"] != nil {

			if ID, IDValid = p.Args["id"].(string); !IDValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if deleteContainerErr = ns.DeleteContainer(namespaces.WithNamespace(context.Background(), namespace), ID); deleteContainerErr != nil {
			return nil, fmt.Errorf("deleteContainer resolver failed to delete %s: %w", ID, deleteContainerErr)
		}

		return nil, nil
	}
}

// NewDeleteTaskResolver returns a graphql resolver that deletes the given task
func NewDeleteTaskResolver(ns node.TaskService) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			namespace, containerID           string
			exitStatus                       node.ExitStatus
			namespaceValid, containerIDValid bool
			deleteTaskErr                    error
		)

		if p.Args["namespace"] != nil {

			if namespace, namespaceValid = p.Args["namespace"].(string); !namespaceValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if p.Args["container_id"] != nil {

			if containerID, containerIDValid = p.Args["container_id"].(string); !containerIDValid {
				return nil, fmt.Errorf("invalid request")
			}
		}

		if exitStatus, deleteTaskErr = ns.DeleteTask(namespaces.WithNamespace(context.Background(), namespace), containerID); deleteTaskErr != nil {
			return nil, fmt.Errorf("deleteTask resolver failed to delete task for %s: %w", containerID, deleteTaskErr)
		}

		return exitStatus, nil
	}
}
