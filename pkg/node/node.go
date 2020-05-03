package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
)

type Node struct {
	Cfg *Config
	Ctr *containerd.Client
}

func NewNode(cfg *Config, ctr *containerd.Client) *Node {
	return &Node{
		Cfg: cfg,
		Ctr: ctr,
	}
}

func (n Node) Serve() (err error) {
	fmt.Println("servin on: " + n.Cfg.Name)
	return nil
}

func (n Node) CreateImage(ctx context.Context, imgRef string) (err error) {
	_, pullErr := n.Ctr.Pull(ctx, imgRef, containerd.WithPullUnpack)

	if pullErr != nil {
		return errors.New("Node failed to pull image " + imgRef + " with error: " + pullErr.Error())
	}

	return nil
}

func (n Node) CreateContainer(ctx context.Context, imgRef, containerName, ssName string) (containerID string, err error) {
	image, getImgErr := n.Ctr.GetImage(ctx, imgRef)

	if getImgErr != nil {
		return "", errors.New("Node failed to get image " + imgRef + " with error: " + getImgErr.Error())
	}

	container, createErr := n.Ctr.NewContainer(
		ctx,
		containerName,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(ssName, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)

	if createErr != nil {
		return "", errors.New("Node failed to create container " + containerName + " with error: " + createErr.Error())
	}

	return container.ID(), nil
}

func (n Node) CreateTask(ctx context.Context, containerID string) (err error) {
	container, loadContainerErr := n.Ctr.LoadContainer(ctx, containerID)

	if loadContainerErr != nil {
		return errors.New("Node failed to load container " + containerID + " with error: " + loadContainerErr.Error())
	}

	_, newTaskErr := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))

	if newTaskErr != nil {
		return errors.New("Node failed to create task with error: " + newTaskErr.Error())
	}

	return nil
}
