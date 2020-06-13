package api

import (
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/node"
)

var imageType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Image",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var containerType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Container",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"image": &graphql.Field{
			Type: imageType,
		},
		"task": &graphql.Field{
			Type: taskType,
		},
	},
})

var taskType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Task",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"container_id": &graphql.Field{
			Type: graphql.Int,
		},
		"pid": &graphql.Field{
			Type: graphql.Int,
		},
		"pids": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
		"status": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// NewImageField creates graphql fields for the image type.
// The image field accepts the arguments defined by the given FieldConfigArgument and is resolved by the function r.
func NewImageField(sp node.ImageService, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        imageType,
		Description: "Get image",
		Args:        args,
		Resolve:     r,
	}
}

// NewImagesField creates graphql fields for the image list type.
// The images field accepts the arguments defined by the given FieldConfigArgument and is resolved by the function r.
func NewImagesField(sp node.ImageService, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        graphql.NewList(imageType),
		Description: "Get image list",
		Args:        args,
		Resolve:     r,
	}
}

// NewContainerField creates graphql fields for the container type.
// The container field accepts the arguments defined by the given FieldConfigArgument and is resolved by the function r.
func NewContainerField(sp node.ContainerService, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        containerType,
		Description: "Get container",
		Args:        args,
		Resolve:     r,
	}
}

// NewContainersField creates graphql fields for the container list type.
// The containers field accepts the arguments defined by the given FieldConfigArgument and is resolved by the function r.
func NewContainersField(sp node.ContainerService, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        graphql.NewList(containerType),
		Description: "Get container list",
		Args:        args,
		Resolve:     r,
	}
}

// NewTaskField creates graphql fields for the task type.
// The task field accepts the arguments defined by the given FieldConfigArgument and is resolved by the function r.
func NewTaskField(sp node.TaskService, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        taskType,
		Description: "Get task",
		Args:        args,
		Resolve:     r,
	}
}

// NewTasksField creates graphql fields for the task list type.
// The tasks field accepts the arguments defined by the given FieldConfigArgument and is resolved by the function r.
func NewTasksField(sp node.TaskService, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        graphql.NewList(taskType),
		Description: "Get task list",
		Args:        args,
		Resolve:     r,
	}
}
