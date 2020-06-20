/*
Package node provides interfaces for containerd interaction and clamor-node business logic.
Right now it's mostly just a wrapper around https://pkg.go.dev/github.com/containerd/containerd?tab=doc#Client.
TODO: Consider k8s container runtime interface support.
*/
package node

import (
	"context"
	"strings"
	"syscall"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/cio"
)

// Node implements the Service interfaces.
type Node struct {
	Ctr *containerd.Client
}

// Service provides core node methods.
type Service interface {
	ImageService
	ContainerService
	TaskService
}

// ImageService provides methods to interact with containerd Image objects.
type ImageService interface {
	PullImage(ctx context.Context, name string) (image Image, err error)
	GetImage(ctx context.Context, name string) (image Image, err error)
	GetImages(ctx context.Context, filter string) (images []Image, err error)
	DeleteImage(ctx context.Context, name string) (err error)
}

// ContainerService provides methods to interact with containerd Container objects.
type ContainerService interface {
	CreateContainer(ctx context.Context, imageName, id string) (container Container, err error)
	GetContainer(ctx context.Context, id string) (container Container, err error)
	GetContainers(ctx context.Context, filter string) (container []Container, err error)
	DeleteContainer(ctx context.Context, id string) (err error)
}

// TaskService provides methods to interact with containerd Task objects.
// Most operations are relative to their parent Container
type TaskService interface {
	CreateTask(ctx context.Context, containerID string) (task Task, err error)
	GetTask(ctx context.Context, containerID string) (task Task, err error)
	GetTasks(ctx context.Context, filter string) (tasks []Task, err error)
	KillTask(ctx context.Context, containerID string) (err error)
	DeleteTask(ctx context.Context, containerID string) (exitStatus ExitStatus, err error)
}

// NewNode returns Node instances.
func NewNode(ctr *containerd.Client) Service {
	return &Node{
		Ctr: ctr,
	}
}

// TaskStatus returns the given containerd.Task's process status as a string.
func TaskStatus(ctx context.Context, task Task) (status string) {
	if task == nil {
		return ""
	}

	stats, statsErr := task.Status(ctx, nil)

	if statsErr != nil {
		return ""
	}

	return string(stats.Status)
}

// GetImage gets a containerd.Image instance by name.
func (n Node) GetImage(ctx context.Context, name string) (i Image, err error) {
	image, err := n.getImage(ctx, name)
	return newImage(image), err
}

// CreateContainer creates a containerd.Container instance with the given id using the given image.
// It returns the created containerd.Container.
func (n Node) CreateContainer(ctx context.Context, imageName, id string) (c Container, err error) {
	var (
		container              containerd.Container
		image                  containerd.Image
		getImageErr, createErr error
	)

	if image, getImageErr = n.getImage(ctx, imageName); getImageErr != nil {
		return nil, fmt.Errorf("failed to get image %s for container %s: %w", imageName, id, getImageErr)
	}

	container, createErr = n.Ctr.NewContainer(
		ctx,
		id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)

	if createErr != nil {
		return nil, fmt.Errorf("failed to create container %s: %w", id, createErr)
	}

	return newContainer(container), nil
}

// CreateTask starts a new task for the given container.
// It returns the created containerd.Task.
// TODO: Figure out task I/O. Currently just using stdio.
func (n Node) CreateTask(ctx context.Context, containerID string) (t Task, err error) {
	var (
		task                        containerd.Task
		container, loadContainerErr = n.Ctr.LoadContainer(ctx, containerID)
		newTaskErr                  error
	)

	if loadContainerErr != nil {
		return nil, fmt.Errorf("failed to load container %s: %w", containerID, loadContainerErr)
	}

	if task, newTaskErr = container.NewTask(ctx, cio.NewCreator(cio.WithStdio)); newTaskErr != nil {
		return nil, fmt.Errorf("failed to create task for container %s: %w", containerID, newTaskErr)
	}

	return newTask(task), nil
}

// PullImage pulls the given image ref.
// It returns the created containerd.Image.
func (n Node) PullImage(ctx context.Context, ref string) (i Image, err error) {
	var img containerd.Image

	if img, err = n.pullImage(ctx, ref); err != nil {
		return nil, err
	}

	return newImage(img), nil
}

// GetImages returns a list of all containerd.Image instances known to the containerd daemon.
func (n Node) GetImages(ctx context.Context, filter string) (images []Image, err error) {
	imgs, getImagesErr := n.Ctr.ListImages(ctx, filter)

	if getImagesErr != nil {
		return nil, fmt.Errorf("failed to get images using filter %s: %w", filter, getImagesErr)
	}

	for _, i := range imgs {
		images = append(images, newImage(i))
	}

	return images, nil
}

// GetContainer retrieves a containerd.Container instance by the given ID.
func (n Node) GetContainer(ctx context.Context, containerID string) (c Container, err error) {
	container, err := n.getContainer(ctx, containerID)
	return newContainer(container), err
}

// GetContainers returns a list of all containerd.Container instances known to the containerd daemon.
func (n Node) GetContainers(ctx context.Context, filter string) (cs []Container, err error) {
	containers, getContainersErr := n.Ctr.Containers(ctx, filter)

	if getContainersErr != nil {
		return nil, fmt.Errorf("failed to get containers using filter %s: %w", filter, getContainersErr)
	}

	for _, c := range containers {
		cs = append(cs, newContainer(c))
	}

	return cs, nil
}

// GetTask retrieves a containerd.Task instance for the given container ID.
// TODO: Handle containers that don't have tasks. The below won't differentiate between errors retrieving a container task and taskless containers.
func (n Node) GetTask(ctx context.Context, containerID string) (t Task, err error) {
	task, err := n.getTask(ctx, containerID)
	return newTask(task), err
}

// GetTasks returns a list of all containerd.Task instances known to the containerd daemon.
func (n Node) GetTasks(ctx context.Context, filter string) (tasks []Task, err error) {
	containers, getContainersErr := n.Ctr.Containers(ctx, filter)

	if getContainersErr != nil {
		return nil, fmt.Errorf("failed to list containers using filter %s: %w", filter, getContainersErr)
	}

	for _, container := range containers {
		task, taskErr := container.Task(ctx, nil)

		if taskErr != nil {
			return nil, fmt.Errorf("failed to build task list: %w", taskErr)
		}

		tasks = append(tasks, newTask(task))
	}

	return tasks, nil
}

// KillTask sends a SIGKILL to the task associated with the given container.
// TODO: Accept other process signals.
func (n Node) KillTask(ctx context.Context, containerID string) (err error) {
	var (
		task                             containerd.Task
		es                               <-chan containerd.ExitStatus
		getTaskErr, waitErr, killTaskErr error
	)

	if task, getTaskErr = n.getTask(ctx, containerID); getTaskErr != nil {
		return fmt.Errorf("failed to get container %s: %w", containerID, getTaskErr)
	}

	if es, waitErr = task.Wait(ctx); waitErr != nil {
		return fmt.Errorf("failed to get task exit status channel: %w", waitErr)
	}

	if killTaskErr = task.Kill(ctx, syscall.SIGKILL); killTaskErr != nil {
		return fmt.Errorf("failed to kill task for container %s: %w", containerID, killTaskErr)
	}

	<-es

	return nil
}

// DeleteImage deletes the given image from the containerd image store.
func (n Node) DeleteImage(ctx context.Context, name string) (err error) {

	if deleteImageErr := n.Ctr.ImageService().Delete(ctx, name); deleteImageErr != nil {
		return fmt.Errorf("failed to delete image %s: %w", name, deleteImageErr)
	}

	return nil
}

// DeleteContainer deletes the given image from the containerd container store.
func (n Node) DeleteContainer(ctx context.Context, id string) (err error) {

	if deleteContainerErr := n.Ctr.ContainerService().Delete(ctx, id); deleteContainerErr != nil {
		return fmt.Errorf("failed to delete container %s: %w", id, deleteContainerErr)
	}

	return nil
}

// DeleteTask deletes resources associated with the given container's task.
func (n Node) DeleteTask(ctx context.Context, containerID string) (exitStatus ExitStatus, err error) {
	var (
		task                      containerd.Task
		taskExitStatus            *containerd.ExitStatus
		getTaskErr, deleteTaskErr error
	)

	if task, getTaskErr = n.getTask(ctx, containerID); getTaskErr != nil {
		return ExitStatus{}, fmt.Errorf("failed to get task for container %s: %w", containerID, getTaskErr)
	}

	if taskExitStatus, deleteTaskErr = task.Delete(ctx); deleteTaskErr != nil {
		return ExitStatus{}, fmt.Errorf("failed to delete task for container %s: %w", containerID, deleteTaskErr)
	}

	return ExitStatus(*taskExitStatus), nil
}

func (n Node) getImage(ctx context.Context, name string) (i containerd.Image, err error) {
	image, err := n.Ctr.GetImage(ctx, name)

	if err == nil {
		return image, nil
	} else if err.Error() == "image \""+name+"\": not found" {
		return nil, ErrNotFound{name: name, inner: err}
	} else {
		return nil, fmt.Errorf("failed to get image %s: %w", name, err)
	}
}

func (n Node) pullImage(ctx context.Context, ref string) (image containerd.Image, err error) {
	image, pullImageErr := n.Ctr.Pull(ctx, ref, containerd.WithPullUnpack)

	if pullImageErr == nil {
		return image, nil
	} else if pullImageErr.Error() == "failed to resolve reference \""+ref+"\": object required" {
		return nil, ErrNotFound{name: ref, inner: pullImageErr}
	} else {
		return nil, fmt.Errorf("failed to pull image %s: %w", ref, pullImageErr)
	}
}

func (n Node) getContainer(ctx context.Context, containerID string) (c containerd.Container, err error) {
	container, loadContainerErr := n.Ctr.LoadContainer(ctx, containerID)

	if loadContainerErr == nil {
		return container, nil
	} else if strings.HasSuffix(loadContainerErr.Error(), "not found") {
		return nil, ErrNotFound{name: containerID, inner: loadContainerErr}
	} else {
		return nil, fmt.Errorf("failed to load container %s: %w", containerID, loadContainerErr)
	}
}

func (n Node) getTask(ctx context.Context, containerID string) (task containerd.Task, err error) {
	container, getContainerErr := n.getContainer(ctx, containerID)

	if getContainerErr != nil {
		return nil, fmt.Errorf("failed to get container for task %s: %w", containerID, getContainerErr)
	}

	task, taskErr := container.Task(ctx, nil)

	if taskErr != nil {
		return nil, fmt.Errorf("failed to load task for container %s: %w", containerID, taskErr)
	}

	return task, nil
}