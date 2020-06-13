package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/containerd/containerd/cio"
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/pkg/node"
	"github.com/mokrz/clamor/pkg/node/api"
)

type image struct {
	name string
}

func NewImage(n string) node.Image {
	return &image{name: n}
}

func (i *image) Name() string {
	return i.name
}

type imageSvc struct {
	images map[string]node.Image
}

func NewImageService(seedImages map[string]node.Image) node.ImageService {
	return &imageSvc{
		images: seedImages,
	}
}

func (is *imageSvc) PullImage(ctx context.Context, name string) (image node.Image, err error) {
	image = NewImage(name)
	is.images[name] = image
	return image, nil
}

func (is *imageSvc) GetImage(ctx context.Context, name string) (image node.Image, err error) {
	var imageValid bool

	if image, imageValid = is.images[name]; !imageValid {
		return nil, errors.New("invalid image")
	}

	return image, nil
}

func (is *imageSvc) GetImages(ctx context.Context, filter string) (images []node.Image, err error) {

	for _, i := range is.images {
		images = append(images, i)
	}

	return images, nil
}

func (is *imageSvc) DeleteImage(ctx context.Context, name string) (err error) {
	delete(is.images, name)
	return nil
}

type container struct {
	id    string
	image node.Image
	task  node.Task
}

func NewContainer(containerID string, i node.Image, t node.Task) node.Container {
	return &container{
		id:    containerID,
		image: i,
		task:  t,
	}
}

func (c *container) ID() string {
	return c.id
}

func (c *container) Image(ctx context.Context) (node.Image, error) {
	return c.image, nil
}

func (c *container) Task(ctx context.Context, attach cio.Attach) (node.Task, error) {
	return c.task, nil
}

type containerService struct {
	containers map[string]node.Container
}

func NewContainerService(seedContainers map[string]node.Container) node.ContainerService {
	return &containerService{
		containers: seedContainers,
	}
}

func (cs *containerService) CreateContainer(ctx context.Context, imageName string, id string) (container node.Container, err error) {
	c := NewContainer(id, NewImage(imageName), NewTask(id, 1, node.Status{}, nil))
	cs.containers[id] = c
	return c, nil
}

func (cs *containerService) GetContainer(ctx context.Context, id string) (container node.Container, err error) {
	var containerValid bool

	if container, containerValid = cs.containers[id]; !containerValid {
		return nil, errors.New("invalid container")
	}

	return container, nil
}

func (cs *containerService) GetContainers(ctx context.Context, filter string) (containers []node.Container, err error) {

	for _, c := range cs.containers {
		containers = append(containers, c)
	}

	return containers, nil
}

func (cs *containerService) DeleteContainer(ctx context.Context, id string) (err error) {
	delete(cs.containers, id)
	return nil
}

type task struct {
	id     string
	pid    uint32
	status node.Status
	pids   []node.ProcessInfo
}

func NewTask(id string, pid uint32, status node.Status, pids []node.ProcessInfo) node.Task {
	return &task{
		id:     id,
		pid:    pid,
		status: status,
		pids:   pids,
	}
}

func (t *task) ID() string {
	return t.id
}

func (t *task) Pid() uint32 {
	return t.pid
}

func (t *task) Status(ctx context.Context, attach cio.Attach) (node.Status, error) {
	return t.status, nil
}

func (t *task) Pids(ctx context.Context) ([]node.ProcessInfo, error) {
	return t.pids, nil
}

type taskService struct {
	tasks map[string]node.Task
}

func NewTaskService(seedTasks map[string]node.Task) node.TaskService {
	return &taskService{
		tasks: seedTasks,
	}
}

func (ts *taskService) CreateTask(ctx context.Context, containerID string) (task node.Task, err error) {
	t := NewTask(containerID, 1, node.Status{}, []node.ProcessInfo{})
	ts.tasks[containerID] = t
	return t, nil
}

func (ts *taskService) GetTask(ctx context.Context, containerID string) (task node.Task, err error) {
	var taskValid bool

	if task, taskValid = ts.tasks[containerID]; !taskValid {
		return nil, errors.New("invalid task")
	}

	return task, nil
}

func (ts *taskService) GetTasks(ctx context.Context, filter string) (tasks []node.Task, err error) {

	for _, t := range ts.tasks {
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (ts *taskService) KillTask(ctx context.Context, containerID string) (err error) {
	return nil
}

func (ts *taskService) DeleteTask(ctx context.Context, containerID string) (exitStatus node.ExitStatus, err error) {
	delete(ts.tasks, containerID)
	return node.ExitStatus{}, nil
}

var (
	weirdString     = "@#%4$1^'`_|+%20"
	testImage       = "docker.io/library/hello-world:latest"
	seedImage       = "docker.io/library/nginx:latest"
	testNamespace   = "clamor-testing"
	testContainerID = "clamor-testing"
)

func TestNewImageResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type imageResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	imageSvc := NewImageService(map[string]node.Image{
		seedImage: NewImage(seedImage),
	})

	tests := []imageResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "ref": testImage}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "ref": testImage}}, wantErr: true},
		{name: "nil ref", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": nil}}, wantErr: true},
		{name: "weird ref", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": weirdString}}, wantErr: true},
		{name: "valid namespace valid ref", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": seedImage}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			imageResolver := api.NewImageResolver(imageSvc)
			i, err := imageResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("image resolver failed with error: " + err.Error())
			}

			if _, imgValid := i.(api.Image); !imgValid && !test.wantErr {
				t.Errorf("image resolver returned incorrect type")
			}
		})
	}
}

func TestNewImagesResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type imageResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	imageSvc := NewImageService(map[string]node.Image{
		seedImage: NewImage(seedImage),
	})
	tests := []imageResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "filter": nil}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "filter": nil}}, wantErr: true},
		{name: "weird filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": weirdString}}, wantErr: true},
		{name: "valid namespace nil filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": nil}}, wantErr: false},
	}

	imageSvc.PullImage(context.TODO(), seedImage)

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			imagesResolver := api.NewImagesResolver(imageSvc)
			imgsRaw, err := imagesResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("image resolver failed with error: " + err.Error())
			}

			imgs, imgsValid := imgsRaw.([]interface{})

			if imgsValid {

				for _, img := range imgs {

					if _, imgValid := img.(api.Image); !imgValid && !test.wantErr {
						t.Errorf("images resolver returned incorrect type")
					}
				}
			}
		})
	}
}

func TestNewContainerResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type imageResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	containerSvc := NewContainerService(map[string]node.Container{
		testContainerID: NewContainer(testContainerID, NewImage(seedImage), NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{})),
	})
	tests := []imageResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "id": testContainerID}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "id": testContainerID}}, wantErr: true},
		{name: "nil container name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": nil}}, wantErr: true},
		{name: "weird container name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": weirdString}}, wantErr: true},
		{name: "valid namespace valid container name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": testContainerID}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			containerResolver := api.NewContainerResolver(containerSvc)
			i, err := containerResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("container resolver failed with error: " + err.Error())
			}

			if _, containerValid := i.(api.Container); !containerValid && !test.wantErr {
				t.Errorf("container resolver returned incorrect type")
			}
		})
	}
}

func TestNewContainersResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type containersResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	containerSvc := NewContainerService(map[string]node.Container{
		testContainerID: NewContainer(testContainerID, NewImage(seedImage), NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{})),
	})
	tests := []containersResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "filter": ""}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "filter": ""}}, wantErr: true},
		{name: "nil filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": nil}}, wantErr: true},
		{name: "weird filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": weirdString}}, wantErr: true},
		{name: "valid namespace valid filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": ""}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			containersResolver := api.NewContainersResolver(containerSvc)
			containersRaw, err := containersResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("containers resolver failed with error: " + err.Error())
			}

			containers, containersValid := containersRaw.([]interface{})

			if containersValid {

				for _, container := range containers {

					if _, containerValid := container.(api.Container); !containerValid && !test.wantErr {
						t.Errorf("containers resolver returned incorrect type")
					}
				}
			}
		})
	}
}

func TestNewTaskResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type taskResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	taskSvc := NewTaskService(map[string]node.Task{
		testContainerID: NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{}),
	})
	tests := []taskResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "container_id": testContainerID}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "container_id": testContainerID}}, wantErr: true},
		{name: "nil container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": nil}}, wantErr: true},
		{name: "weird container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": weirdString}}, wantErr: true},
		{name: "valid namespace valid container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": testContainerID}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			taskResolver := api.NewTaskResolver(taskSvc)
			task, err := taskResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("task resolver failed with error: " + err.Error())
			}

			if _, taskValid := task.(api.Task); !taskValid && !test.wantErr {
				t.Errorf("task resolver returned incorrect type")
			}
		})
	}
}

func TestNewTasksResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type tasksResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	taskSvc := NewTaskService(map[string]node.Task{
		testContainerID: NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{}),
	})
	tests := []tasksResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "filter": ""}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "filter": ""}}, wantErr: true},
		{name: "nil filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": nil}}, wantErr: true},
		{name: "weird filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": weirdString}}, wantErr: true},
		{name: "valid namespace valid filter", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "filter": ""}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			tasksResolver := api.NewTasksResolver(taskSvc)
			tasksRaw, err := tasksResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("tasks resolver failed with error: " + err.Error())
			}

			tasks, tasksValid := tasksRaw.([]interface{})

			if tasksValid {

				for _, task := range tasks {

					if _, taskValid := task.(api.Task); !taskValid && !test.wantErr {
						t.Errorf("tasks resolver returned incorrect type")
					}
				}
			}
		})
	}
}

func TestNewCreateImageResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type imageResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	imageSvc := NewImageService(map[string]node.Image{
		seedImage: NewImage(seedImage),
	})

	tests := []imageResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "ref": testImage}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "ref": testImage}}, wantErr: true},
		{name: "nil ref", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": nil}}, wantErr: true},
		{name: "weird ref", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": weirdString}}, wantErr: true},
		{name: "valid namespace valid ref", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": seedImage}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			createImageResolver := api.NewCreateImageResolver(imageSvc)
			i, err := createImageResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("create image resolver failed with error: " + err.Error())
			}

			if _, imgValid := i.(api.Image); !imgValid && !test.wantErr {
				t.Errorf("create image resolver returned incorrect type")
			}
		})
	}
}

func TestNewCreateContainerResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type createContainerResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	containerSvc := NewContainerService(map[string]node.Container{
		testContainerID: NewContainer(testContainerID, NewImage(seedImage), NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{})),
	})
	tests := []createContainerResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "id": testContainerID, "image_name": testImage}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "id": testContainerID, "image_name": testImage}}, wantErr: true},
		{name: "nil container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": nil, "image_name": testImage}}, wantErr: true},
		{name: "weird container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": weirdString, "image_name": testImage}}, wantErr: true},
		{name: "nil image name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": testContainerID, "image_name": nil}}, wantErr: true},
		{name: "weird image name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": testContainerID, "image_name": weirdString}}, wantErr: true},
		{name: "valid namespace valid container ID valid image name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": testContainerID, "image_name": testImage}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			containerResolver := api.NewCreateContainerResolver(containerSvc)
			i, err := containerResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("create container resolver failed with error: " + err.Error())
			}

			if _, containerValid := i.(api.Container); !containerValid && !test.wantErr {
				t.Errorf("create container resolver returned incorrect type")
			}
		})
	}
}

func TestNewCreateTaskResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type createTaskResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	taskSvc := NewTaskService(map[string]node.Task{
		testContainerID: NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{}),
	})
	tests := []createTaskResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "container_id": testContainerID}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "container_id": testContainerID}}, wantErr: true},
		{name: "nil container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": nil}}, wantErr: true},
		{name: "weird container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": weirdString}}, wantErr: true},
		{name: "valid namespace valid container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": testContainerID}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			createTaskResolver := api.NewCreateTaskResolver(taskSvc)
			i, err := createTaskResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("create task resolver failed with error: " + err.Error())
			}

			if _, taskValid := i.(api.Task); !taskValid && !test.wantErr {
				t.Errorf("create task resolver returned incorrect type")
			}
		})
	}
}

func TestNewKillTaskResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type killTaskResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	taskSvc := NewTaskService(map[string]node.Task{
		testContainerID: NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{}),
	})
	tests := []killTaskResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "container_id": testContainerID}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "container_id": testContainerID}}, wantErr: true},
		{name: "nil container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": nil}}, wantErr: true},
		{name: "weird container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": weirdString}}, wantErr: true},
		{name: "valid namespace valid container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": testContainerID}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			killTaskResolver := api.NewKillTaskResolver(taskSvc)
			_, err := killTaskResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("kill task resolver failed with error: " + err.Error())
			}
		})
	}
}

func TestNewDeleteTaskResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type deleteTaskResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	taskSvc := NewTaskService(map[string]node.Task{
		testContainerID: NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{}),
	})
	tests := []deleteTaskResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "container_id": testContainerID}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "container_id": testContainerID}}, wantErr: true},
		{name: "nil container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": nil}}, wantErr: true},
		{name: "weird container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": weirdString}}, wantErr: true},
		{name: "valid namespace valid container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "container_id": testContainerID}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			deleteTaskResolver := api.NewDeleteTaskResolver(taskSvc)
			_, err := deleteTaskResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("delete task resolver failed with error: " + err.Error())
			}
		})
	}
}

func TestNewDeleteContainerResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type deleteContainerResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	containerSvc := NewContainerService(map[string]node.Container{
		testContainerID: NewContainer(testContainerID, NewImage(seedImage), NewTask(testContainerID, 1, node.Status{}, []node.ProcessInfo{})),
	})
	tests := []deleteContainerResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "id": testContainerID}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "id": testContainerID}}, wantErr: true},
		{name: "nil container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": nil}}, wantErr: true},
		{name: "weird container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": weirdString}}, wantErr: true},
		{name: "valid namespace valid container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "id": testContainerID}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			deleteContainerResolver := api.NewDeleteContainerResolver(containerSvc)
			_, err := deleteContainerResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("delete container resolver failed with error: " + err.Error())
			}
		})
	}
}

func TestNewDeleteImageResolver(t *testing.T) {
	type resolverArgs struct {
		resolveParamArgs map[string]interface{}
	}

	type deleteContainerResolverTest struct {
		name    string
		args    resolverArgs
		wantErr bool
	}

	imageSvc := NewImageService(map[string]node.Image{
		seedImage: NewImage(seedImage),
	})
	tests := []deleteContainerResolverTest{
		{name: "empty namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": "", "ref": seedImage}}, wantErr: true},
		{name: "weird namespace", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": weirdString, "ref": seedImage}}, wantErr: true},
		{name: "nil image name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": nil}}, wantErr: true},
		{name: "weird image name", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": weirdString}}, wantErr: true},
		{name: "valid namespace valid container ID", args: resolverArgs{resolveParamArgs: map[string]interface{}{"namespace": testNamespace, "ref": seedImage}}, wantErr: false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			deleteImageResolver := api.NewDeleteImageResolver(imageSvc)
			_, err := deleteImageResolver(graphql.ResolveParams{
				Args: test.args.resolveParamArgs,
			})

			if err != nil && !test.wantErr {
				t.Errorf("delete image resolver failed with error: " + err.Error())
			}
		})
	}
}
