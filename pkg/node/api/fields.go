package api

import (
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/pkg/node"
)

var imageType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Image",
	Fields: graphql.Fields{
		"ref": &graphql.Field{
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

func newImageField(sp node.Service, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        imageType,
		Description: "Get image by ref",
		Args:        args,
		Resolve:     r,
	}
}

func newImagesField(sp node.Service, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        graphql.NewList(imageType),
		Description: "Get images",
		Args:        args,
		Resolve:     r,
	}
}

func newContainerField(sp node.Service, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        containerType,
		Description: "Get container by ID",
		Args:        args,
		Resolve:     r,
	}
}

func newContainersField(sp node.Service, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        graphql.NewList(containerType),
		Description: "Get image by ref",
		Args:        args,
		Resolve:     r,
	}
}

func newTaskField(sp node.Service, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        taskType,
		Description: "Get image by ref",
		Args:        args,
		Resolve:     r,
	}
}

func newTasksField(sp node.Service, r graphql.FieldResolveFn, args graphql.FieldConfigArgument) (field *graphql.Field) {
	return &graphql.Field{
		Type:        graphql.NewList(taskType),
		Description: "Get image by ref",
		Args:        args,
		Resolve:     r,
	}
}
