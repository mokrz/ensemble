package node_test

import (
	"context"
	"fmt"
	"math/rand"
	"syscall"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/mokrz/clamor/node"
)

var (
	containerdSock  = "/run/containerd/containerd.sock"
	weirdString     = "@#%4$1^'`_|+%20"
	testImage       = "docker.io/library/hello-world:latest"
	testNamespace   = "clamor-testing"
	testContainerID = "clamor-testing"
	ctr             *containerd.Client
)

func TestPullImage(t *testing.T) {
	type testArguments struct {
		namespace, imageName string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", imageName: testImage}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, imageName: testImage}, wantErr: true},
		{name: "empty image name", args: testArguments{namespace: testNamespace, imageName: ""}, wantErr: true},
		{name: "weird image name", args: testArguments{namespace: testNamespace, imageName: weirdString}, wantErr: true},
		{name: "valid namespace valid image name", args: testArguments{namespace: testNamespace, imageName: testImage}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := namespaces.WithNamespace(context.TODO(), test.args.namespace)
			i, err := node.PullImage(ctx, test.args.imageName)

			if i != nil {
				ctrd.deleteImage(ctx, i.Name())
			} else if !test.wantErr {
				t.Errorf("node.PullImage failed with error: %s", err.Error())
			}
		})
	}
}

func TestGetImage(t *testing.T) {
	type testArguments struct {
		namespace, imageName string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", imageName: testImage}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, imageName: testImage}, wantErr: true},
		{name: "empty image name", args: testArguments{namespace: testNamespace, imageName: ""}, wantErr: true},
		{name: "weird image name", args: testArguments{namespace: testNamespace, imageName: weirdString}, wantErr: true},
		{name: "valid namespace valid image name", args: testArguments{namespace: testNamespace, imageName: testImage}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)
	ctx := namespaces.WithNamespace(context.TODO(), testNamespace)

	_, pullErr := ctrd.pullImage(ctx, testImage)

	if pullErr != nil {
		t.Errorf("failed to pull seed image with error: %s", pullErr)
	}

	defer ctrd.deleteImage(ctx, testImage)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := node.GetImage(namespaces.WithNamespace(context.TODO(), test.args.namespace), test.args.imageName)

			if err != nil && !test.wantErr {
				t.Errorf("node.GetImage failed with error: %s", err.Error())
			}
		})
	}
}

func TestCreateContainer(t *testing.T) {
	type testArguments struct {
		namespace, image, id string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", image: testImage, id: testContainerID}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, image: testImage, id: testContainerID}, wantErr: true},
		{name: "empty image name", args: testArguments{namespace: testNamespace, image: "", id: ""}, wantErr: true},
		{name: "weird image name", args: testArguments{namespace: testNamespace, image: weirdString, id: weirdString}, wantErr: true},
		{name: "empty container ID", args: testArguments{namespace: testNamespace, image: testImage, id: ""}, wantErr: true},
		{name: "weird container ID", args: testArguments{namespace: testNamespace, image: testImage, id: weirdString}, wantErr: true},
		{name: "valid namespace valid image valid container ID", args: testArguments{namespace: testNamespace, image: testImage, id: testContainerID}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)

	ctx := namespaces.WithNamespace(context.TODO(), testNamespace)

	_, pullErr := ctrd.pullImage(ctx, testImage)

	if pullErr != nil {
		t.Errorf("failed to pull seed image with error: %s", pullErr.Error())
	}

	defer ctrd.deleteImage(ctx, testImage)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := namespaces.WithNamespace(context.TODO(), test.args.namespace)

			_, err := node.CreateContainer(ctx, test.args.image, test.args.id)

			if err != nil && !test.wantErr {
				t.Errorf("node.CreateContainer failed with error: %s", err.Error())
			}

			ctrd.deleteContainer(ctx, test.args.id)
		})
	}
}

func TestCreateTask(t *testing.T) {
	type testArguments struct {
		namespace, containerID string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", containerID: testContainerID}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, containerID: testContainerID}, wantErr: true},
		{name: "empty container ID", args: testArguments{namespace: testNamespace, containerID: ""}, wantErr: true},
		{name: "weird container ID", args: testArguments{namespace: testNamespace, containerID: weirdString}, wantErr: true},
		{name: "valid namespace valid container ID", args: testArguments{namespace: testNamespace, containerID: testContainerID}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)

	ctx := namespaces.WithNamespace(context.TODO(), testNamespace)

	_, pullErr := ctrd.pullImage(ctx, testImage)

	if pullErr != nil {
		t.Errorf("failed to pull seed image with error: %s", pullErr.Error())
	}

	defer ctrd.deleteImage(ctx, testImage)

	_, createContainerErr := ctrd.createContainer(namespaces.WithNamespace(context.TODO(), testNamespace), testImage, testContainerID)

	if createContainerErr != nil {
		t.Errorf("failed to create seed container with error: %s", createContainerErr.Error())
	}

	defer ctrd.deleteContainer(ctx, testContainerID)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := namespaces.WithNamespace(context.TODO(), test.args.namespace)

			if tsk, err := node.CreateTask(ctx, test.args.containerID); err != nil && !test.wantErr {
				t.Errorf("node.CreateTask failed with error: %s", err.Error())
			} else if tsk != nil {
				ctrd.killTask(ctx, tsk.ID())
				ctrd.deleteTask(ctx, tsk.ID())
			}
		})
	}
}

func TestKillTask(t *testing.T) {
	type testArguments struct {
		namespace, containerID string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", containerID: testContainerID}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, containerID: testContainerID}, wantErr: true},
		{name: "empty container ID", args: testArguments{namespace: testNamespace, containerID: ""}, wantErr: true},
		{name: "weird container ID", args: testArguments{namespace: testNamespace, containerID: weirdString}, wantErr: true},
		{name: "valid namespace valid container ID", args: testArguments{namespace: testNamespace, containerID: testContainerID}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)

	ctx := namespaces.WithNamespace(context.TODO(), testNamespace)

	_, pullErr := ctrd.pullImage(ctx, testImage)

	if pullErr != nil {
		t.Errorf("failed to pull seed image with error: %s", pullErr.Error())
	}

	defer ctrd.deleteImage(ctx, testImage)

	_, createContainerErr := ctrd.createContainer(namespaces.WithNamespace(context.TODO(), testNamespace), testImage, testContainerID)

	if createContainerErr != nil {
		t.Errorf("failed to create seed container with error: %s", createContainerErr.Error())
	}

	defer ctrd.deleteContainer(ctx, testContainerID)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrd.createTask(ctx, test.args.containerID)
			err := node.KillTask(namespaces.WithNamespace(context.TODO(), test.args.namespace), test.args.containerID)

			if err == nil {
				ctrd.deleteTask(ctx, test.args.containerID)
			} else if !test.wantErr {
				t.Errorf("node.KillTask failed with error: %s", err.Error())
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	type testArguments struct {
		namespace, containerID string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	containerID := testContainerID + randString(8)
	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", containerID: containerID}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, containerID: containerID}, wantErr: true},
		{name: "empty container ID", args: testArguments{namespace: testNamespace, containerID: ""}, wantErr: true},
		{name: "weird container ID", args: testArguments{namespace: testNamespace, containerID: weirdString}, wantErr: true},
		{name: "valid namespace valid container ID", args: testArguments{namespace: testNamespace, containerID: containerID}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)
	ctx := namespaces.WithNamespace(context.TODO(), testNamespace)

	_, createContainerErr := ctrd.createContainer(ctx, testImage, containerID)

	if createContainerErr != nil {
		t.Errorf("failed to create container with error: %s", createContainerErr.Error())
	}

	defer ctrd.deleteContainer(ctx, containerID)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := namespaces.WithNamespace(context.TODO(), test.args.namespace)
			ctrd.createTask(ctx, containerID)
			ctrd.killTask(ctx, containerID)
			_, err := node.DeleteTask(ctx, test.args.containerID)

			if err != nil && !test.wantErr {
				t.Errorf("node.KillTask failed with error: %s", err.Error())
			}
		})
	}
}

func TestDeleteContainer(t *testing.T) {
	type testArguments struct {
		namespace, containerID string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", containerID: testContainerID}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, containerID: testContainerID}, wantErr: true},
		{name: "empty container ID", args: testArguments{namespace: testNamespace, containerID: ""}, wantErr: true},
		{name: "weird container ID", args: testArguments{namespace: testNamespace, containerID: weirdString}, wantErr: true},
		{name: "valid namespace valid container ID", args: testArguments{namespace: testNamespace, containerID: testContainerID}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := namespaces.WithNamespace(context.TODO(), test.args.namespace)
			ctrd.createContainer(ctx, testImage, testContainerID)
			err := node.DeleteContainer(ctx, test.args.containerID)

			if err != nil && !test.wantErr {
				t.Errorf("node.DeleteContainer failed with error: %s", err.Error())
			}
		})
	}
}

func TestDeleteImage(t *testing.T) {
	type testArguments struct {
		namespace, imageName string
	}

	type test struct {
		name             string
		args             testArguments
		wantErr, wantImg bool
	}

	tests := []test{
		{name: "empty namespace", args: testArguments{namespace: "", imageName: testImage}, wantErr: true},
		{name: "weird namespace", args: testArguments{namespace: weirdString, imageName: testImage}, wantErr: true},
		{name: "empty image name", args: testArguments{namespace: testNamespace, imageName: ""}, wantErr: true},
		{name: "weird image name", args: testArguments{namespace: testNamespace, imageName: weirdString}, wantErr: true},
		{name: "valid namespace valid image name", args: testArguments{namespace: testNamespace, imageName: testImage}, wantErr: false},
	}

	ctrd, ctrdErr := newCtrd(containerdSock)
	defer ctrd.client.Close()

	if ctrdErr != nil {
		t.Errorf("failed to create containerd client with error: %s", ctrdErr.Error())
	}

	node := node.NewNode(ctrd.client)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := namespaces.WithNamespace(context.TODO(), test.args.namespace)
			ctrd.pullImage(ctx, testImage)

			err := node.DeleteImage(ctx, test.args.imageName)

			if err != nil && !test.wantErr {
				t.Errorf("node.DeleteImage failed with error: %s", err.Error())
			}
		})
	}
}

func randString(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

// ctrd is similar to the containerd interaction provided by node.Service methods, but it's meant to be a simpler implementation that's easier to trust.
// node.Service tests should use these methods to create SUT dependencies instead of whatever their closest node.Service relative is.
// e.g. A test for node.CreateTask could use the below ctrd.createContainer method to populate the containerd daemon with the required container prior to task creation.

type ctrd struct {
	client *containerd.Client
}

func newCtrd(containerdPath string) (*ctrd, error) {
	var (
		ctr    *containerd.Client
		ctrErr error
	)

	if ctr, ctrErr = containerd.New(containerdPath); ctrErr != nil {
		return nil, ctrErr
	}

	return &ctrd{
		client: ctr,
	}, nil
}

func (c *ctrd) pullImage(ctx context.Context, imageName string) (containerd.Image, error) {
	var (
		image        containerd.Image
		pullImageErr error
	)

	image, pullImageErr = c.client.Pull(ctx, imageName, containerd.WithPullUnpack)

	if pullImageErr != nil {
		return nil, fmt.Errorf("failed to pull image ref %s with error: %s", imageName, pullImageErr.Error())
	}

	return image, nil
}

func (c *ctrd) createContainer(ctx context.Context, imageName, id string) (containerd.Container, error) {
	var (
		image   containerd.Image
		pullErr error
	)

	if image, pullErr = c.client.Pull(ctx, imageName, containerd.WithPullUnpack); pullErr != nil {
		return nil, fmt.Errorf("failed to pull image ref %s with error: %s", imageName, pullErr.Error())
	}

	return c.client.NewContainer(
		ctx,
		id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)
}

func (c *ctrd) getContainer(ctx context.Context, id string) (containerd.Container, error) {
	container, loadContainerErr := c.client.LoadContainer(ctx, id)

	if loadContainerErr != nil {
		return nil, fmt.Errorf("failed to load container %s with error: %s", id, loadContainerErr.Error())
	}

	return container, nil
}

func (c *ctrd) deleteContainer(ctx context.Context, id string) error {
	var (
		container    containerd.Container
		containerErr error
	)

	if container, containerErr = c.client.LoadContainer(ctx, id); containerErr != nil {
		return fmt.Errorf("failed to load container with error: %s", containerErr.Error())
	}

	return container.Delete(ctx, containerd.WithSnapshotCleanup)
}

func (c *ctrd) createTask(ctx context.Context, containerID string) (containerd.Task, error) {
	var (
		task                         containerd.Task
		container                    containerd.Container
		loadContainerErr, newTaskErr error
	)

	if container, loadContainerErr = c.client.LoadContainer(ctx, containerID); loadContainerErr != nil {
		return nil, fmt.Errorf("Node failed to load container %s with error: %s", containerID, loadContainerErr.Error())
	}

	if task, newTaskErr = container.NewTask(ctx, cio.NewCreator(cio.WithStdio)); newTaskErr != nil {
		return nil, fmt.Errorf("Node failed to create task with error: %s", newTaskErr.Error())
	}

	return task, nil
}

func (c *ctrd) killTask(ctx context.Context, containerID string) error {
	var (
		task                             containerd.Task
		es                               <-chan containerd.ExitStatus
		getTaskErr, waitErr, killTaskErr error
	)

	if task, getTaskErr = c.getTask(ctx, containerID); getTaskErr != nil {
		return fmt.Errorf("failed to get container with error: %s", getTaskErr.Error())
	}

	if es, waitErr = task.Wait(ctx); waitErr != nil {
		return fmt.Errorf("failed get task exit status channel with error: %s", waitErr.Error())
	}

	if killTaskErr = task.Kill(ctx, syscall.SIGKILL); killTaskErr != nil {
		return fmt.Errorf("failed to kill task for container %s with error: %s", containerID, killTaskErr.Error())
	}

	<-es

	return killTaskErr
}

func (c *ctrd) deleteTask(ctx context.Context, containerID string) error {
	var (
		container                               containerd.Container
		task                                    containerd.Task
		getContainerErr, taskErr, deleteTaskErr error
	)

	if container, getContainerErr = c.client.LoadContainer(ctx, containerID); getContainerErr != nil {
		return fmt.Errorf("failed to load task for container %s with error: %s", containerID, getContainerErr.Error())
	}

	if task, taskErr = container.Task(ctx, nil); taskErr != nil {
		return fmt.Errorf("failed to load task for container %s with error: %s", containerID, taskErr.Error())
	}

	if _, deleteTaskErr = task.Delete(ctx); deleteTaskErr != nil {
		return fmt.Errorf("failed to delete task for container %s with error: %s", containerID, deleteTaskErr.Error())
	}

	return nil
}

func (c *ctrd) deleteImage(ctx context.Context, name string) error {
	var (
		getImageErr, deleteImageErr error
	)

	if deleteImageErr = c.client.ImageService().Delete(ctx, name); deleteImageErr != nil {
		return fmt.Errorf("failed to get image with error: %s", getImageErr.Error())
	}

	return nil
}

func (c *ctrd) getTask(ctx context.Context, containerID string) (containerd.Task, error) {
	container, getContainerErr := c.getContainer(ctx, containerID)

	if getContainerErr != nil {
		return nil, fmt.Errorf("failed to load container for task %s: %w", containerID, getContainerErr)
	}

	task, taskErr := container.Task(ctx, nil)

	if taskErr != nil {
		return nil, fmt.Errorf("failed to load task for container %s: %w", containerID, taskErr)
	}

	return task, nil
}
