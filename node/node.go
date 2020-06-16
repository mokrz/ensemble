/*
Package node provides interfaces for containerd interaction and clamor-node business logic.
Right now it's mostly just a wrapper around https://pkg.go.dev/github.com/containerd/containerd?tab=doc#Client.
TODO: Consider k8s container runtime interface support.
*/
package node

import (
	"context"
	"errors"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
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
func NewNode(ctr *containerd.Client) *Node {
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

func (n Node) getImage(ctx context.Context, name string) (i containerd.Image, err error) {
	image, err := n.Ctr.GetImage(ctx, name)

	if err != nil {
		return nil, errors.New("GetImage: failed to get image " + name + " with error: " + err.Error())
	}

	return image, nil
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
		return nil, errors.New("Node failed to get image " + imageName + " with error: " + getImageErr.Error())
	}

	container, createErr = n.Ctr.NewContainer(
		ctx,
		id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)

	if createErr != nil {
		return nil, errors.New("Node failed to create container " + id + " with error: " + createErr.Error())
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
		return nil, errors.New("Node failed to load container " + containerID + " with error: " + loadContainerErr.Error())
	}

	if task, newTaskErr = container.NewTask(ctx, cio.NewCreator(cio.WithStdio)); newTaskErr != nil {
		return nil, errors.New("Node failed to create task with error: " + newTaskErr.Error())
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

func (n Node) pullImage(ctx context.Context, ref string) (image containerd.Image, err error) {
	image, pullImageErr := n.Ctr.Pull(ctx, ref, containerd.WithPullUnpack)

	if pullImageErr != nil {
		return nil, errors.New("PullImage: failed to pull image ref " + ref + " with error: " + pullImageErr.Error())
	}

	return image, nil
}

// GetImages returns a list of all containerd.Image instances known to the containerd daemon.
func (n Node) GetImages(ctx context.Context, filter string) (images []Image, err error) {
	imgs, getImagesErr := n.Ctr.ListImages(ctx, filter)

	if getImagesErr != nil {
		return nil, errors.New("GetImages: failed to get images using filter " + filter + " with error: " + getImagesErr.Error())
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

func (n Node) getContainer(ctx context.Context, containerID string) (c containerd.Container, err error) {
	container, loadContainerErr := n.Ctr.LoadContainer(ctx, containerID)

	if loadContainerErr != nil {
		return nil, errors.New("GetContainer: failed to load container " + containerID + " with error: " + loadContainerErr.Error())
	}

	return container, nil
}

// GetContainers returns a list of all containerd.Container instances known to the containerd daemon.
func (n Node) GetContainers(ctx context.Context, filter string) (cs []Container, err error) {
	containers, getContainersErr := n.Ctr.Containers(ctx, filter)

	if getContainersErr != nil {
		return nil, errors.New("GetImages: failed to get containers using filter " + filter + " with error: " + getContainersErr.Error())
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

func (n Node) getTask(ctx context.Context, containerID string) (task containerd.Task, err error) {
	container, getContainerErr := n.getContainer(ctx, containerID)

	if getContainerErr != nil {
		return nil, errors.New("GetTask: failed to load task for container " + containerID + " with error: " + getContainerErr.Error())
	}

	task, taskErr := container.Task(ctx, nil)

	if taskErr != nil {
		return nil, errors.New("GetTask: failed to load task for container " + containerID + " with error: " + taskErr.Error())
	}

	return task, nil
}

// GetTasks returns a list of all containerd.Task instances known to the containerd daemon.
func (n Node) GetTasks(ctx context.Context, filter string) (tasks []Task, err error) {
	containers, getContainersErr := n.Ctr.Containers(ctx, filter)

	if getContainersErr != nil {
		return nil, errors.New("GetTask: failed to list containers using filter " + filter + " with error: " + getContainersErr.Error())
	}

	for _, container := range containers {
		task, taskErr := container.Task(ctx, nil)

		if taskErr != nil {
			//TODO: Log error
			continue
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
		return errors.New("KillTask: failed to get container " + containerID + " with error: " + getTaskErr.Error())
	}
	
	if es, waitErr = task.Wait(ctx); waitErr != nil {
		return errors.New("KillTask: failed get task exit status channel with error:" + waitErr.Error())
	}

	if killTaskErr = task.Kill(ctx, syscall.SIGKILL); killTaskErr != nil {
		return errors.New("KillTask: failed to kill task for container " + containerID + " with error: " + killTaskErr.Error())
	}
	
	<-es

	return nil
}

// DeleteImage deletes the given image from the containerd image store.
func (n Node) DeleteImage(ctx context.Context, name string) (err error) {

	if deleteImageErr := n.Ctr.ImageService().Delete(ctx, name); deleteImageErr != nil {
		return errors.New("DeleteImage: failed to delete image " + name + " with error: " + deleteImageErr.Error())
	}

	return nil
}

// DeleteContainer deletes the given image from the containerd container store.
func (n Node) DeleteContainer(ctx context.Context, id string) (err error) {

	if deleteContainerErr := n.Ctr.ContainerService().Delete(ctx, id); deleteContainerErr != nil {
		return errors.New("DeleteContainer: failed to delete container " + id + " with error: " + deleteContainerErr.Error())
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
		return ExitStatus{}, errors.New("DeleteTask: failed to get task for container " + containerID + " with error: " + getTaskErr.Error())
	}

	if taskExitStatus, deleteTaskErr = task.Delete(ctx); deleteTaskErr != nil {
		return ExitStatus{}, errors.New("DeleteTask: failed to delete task for container " + containerID + " with error: " + deleteTaskErr.Error())
	}

	return ExitStatus(*taskExitStatus), nil
}
