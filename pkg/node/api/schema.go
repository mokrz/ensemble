package api

import (
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/pkg/node"
)

// NewGraphQLSchema returns a new graphql schema instance containing root Query and root Mutation types.
// It's responsible for allocating the remainder of the API's graphql fields and wiring them to their respective resolvers + arguments.
func NewGraphQLSchema(ns node.Service) (schema graphql.Schema, err error) {
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"image":      newImageField(ns, NewImageResolver(ns), imageArgs),
			"images":     newImagesField(ns, NewImagesResolver(ns), imagesArgs),
			"container":  newContainerField(ns, NewContainerResolver(ns), containerArgs),
			"containers": newContainersField(ns, NewContainersResolver(ns), containersArgs),
			"task":       newTaskField(ns, NewTaskResolver(ns), taskArgs),
			"tasks":      newTasksField(ns, NewTasksResolver(ns), tasksArgs),
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createImage":     newImageField(ns, NewCreateImageResolver(ns), createImageArgs),
			"createContainer": newContainerField(ns, NewCreateContainerResolver(ns), createContainerArgs),
			"createTask":      newTaskField(ns, NewCreateTaskResolver(ns), createTaskArgs),
			"deleteImage":     newImageField(ns, NewDeleteImageResolver(ns), imageArgs),
			"deleteContainer": newContainerField(ns, NewDeleteContainerResolver(ns), containerArgs),
			"deleteTask":      newTaskField(ns, NewDeleteTaskResolver(ns), taskArgs),
			"killTask":        newTaskField(ns, NewKillTaskResolver(ns), taskArgs),
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: queryType, Mutation: mutationType})
}
