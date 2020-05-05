package node

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	grpc "google.golang.org/grpc"
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
	listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", n.Cfg.APIHost, n.Cfg.APIPort))

	if listenErr != nil {
		return errors.New("Node failed to create an API listener on " + n.Cfg.APIHost + ":" + strconv.Itoa(n.Cfg.APIPort) + " with error: " + listenErr.Error())
	}

	grpcServer := grpc.NewServer()
	RegisterNodeServiceServer(grpcServer, n)
	serveErr := grpcServer.Serve(listener)

	if serveErr != nil {
		return errors.New("Node API failed to serve on " + n.Cfg.APIHost + ":" + strconv.Itoa(n.Cfg.APIPort) + " with error: " + serveErr.Error())
	}

	return nil
}

func (n Node) CreateImage(ctx context.Context, req *CreateImageRequest) (*CreateImageResponse, error) {
	_, pullErr := n.Ctr.Pull(ctx, req.Ref, containerd.WithPullUnpack)

	if pullErr != nil {
		return nil, errors.New("Node failed to pull image " + req.Ref + " with error: " + pullErr.Error())
	}

	return nil, nil
}

func (n Node) CreateContainer(ctx context.Context, req *CreateContainerRequest) (*CreateContainerResponse, error) {
	image, getImgErr := n.Ctr.GetImage(ctx, req.ImageRef)

	if getImgErr != nil {
		return nil, errors.New("Node failed to get image " + req.ImageRef + " with error: " + getImgErr.Error())
	}

	container, createErr := n.Ctr.NewContainer(
		ctx,
		req.ContainerName,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(req.ContainerName, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)

	if createErr != nil {
		return nil, errors.New("Node failed to create container " + req.ContainerName + " with error: " + createErr.Error())
	}

	return &CreateContainerResponse{ContainerID: container.ID()}, nil
}

func (n Node) CreateTask(ctx context.Context, req *CreateTaskRequest) (*CreateTaskResponse, error) {
	container, loadContainerErr := n.Ctr.LoadContainer(ctx, req.ContainerID)

	if loadContainerErr != nil {
		return nil, errors.New("Node failed to load container " + req.ContainerID + " with error: " + loadContainerErr.Error())
	}

	_, newTaskErr := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))

	if newTaskErr != nil {
		return nil, errors.New("Node failed to create task with error: " + newTaskErr.Error())
	}

	return nil, nil
}
