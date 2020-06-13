package api

import (
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/node"
)

// NewGraphQLSchema returns a new graphql schema instance containing root Query and root Mutation types.
// It's responsible for allocating the remainder of the API's graphql fields and wiring them to their respective resolvers + arguments.
func NewGraphQLSchema(ns node.Service) (schema graphql.Schema, err error) {
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"image":      NewImageField(ns, NewImageResolver(ns), imageArgs),
			"images":     NewImagesField(ns, NewImagesResolver(ns), imagesArgs),
			"container":  NewContainerField(ns, NewContainerResolver(ns), containerArgs),
			"containers": NewContainersField(ns, NewContainersResolver(ns), containersArgs),
			"task":       NewTaskField(ns, NewTaskResolver(ns), taskArgs),
			"tasks":      NewTasksField(ns, NewTasksResolver(ns), tasksArgs),
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createImage":     NewImageField(ns, NewCreateImageResolver(ns), createImageArgs),
			"createContainer": NewContainerField(ns, NewCreateContainerResolver(ns), createContainerArgs),
			"createTask":      NewTaskField(ns, NewCreateTaskResolver(ns), createTaskArgs),
			"deleteImage":     NewImageField(ns, NewDeleteImageResolver(ns), imageArgs),
			"deleteContainer": NewContainerField(ns, NewDeleteContainerResolver(ns), containerArgs),
			"deleteTask":      NewTaskField(ns, NewDeleteTaskResolver(ns), taskArgs),
			"killTask":        NewTaskField(ns, NewKillTaskResolver(ns), taskArgs),
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: queryType, Mutation: mutationType})
}
