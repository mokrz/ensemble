package api

import (
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/pkg/node"
)

// NewGraphQLSchema returns a new graphql.Schema instance containing root Query and root Mutation types.
// It's responsible for allocating the remainder of the API's graphql fields and wiring them to their resolvers/arguments.
func NewGraphQLSchema(sp node.Service) (schema graphql.Schema, err error) {
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"image":      newImageField(sp, newImageResolver(sp), imageArgs),
			"images":     newImagesField(sp, newImagesResolver(sp), imagesArgs),
			"container":  newContainerField(sp, newContainerResolver(sp), containerArgs),
			"containers": newContainersField(sp, newContainersResolver(sp), containersArgs),
			"task":       newTaskField(sp, newTaskResolver(sp), taskArgs),
			"tasks":      newTasksField(sp, newTasksResolver(sp), tasksArgs),
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createImage":     newImageField(sp, newCreateImageResolver(sp), createImageArgs),
			"createContainer": newContainerField(sp, newCreateContainerResolver(sp), createContainerArgs),
			"createTask":      newTaskField(sp, newCreateTaskResolver(sp), createTaskArgs),
			"deleteImage":     newImageField(sp, newDeleteImageResolver(sp), imageArgs),
			"deleteContainer": newContainerField(sp, newDeleteContainerResolver(sp), containerArgs),
			"deleteTask":      newTaskField(sp, newDeleteTaskResolver(sp), taskArgs),
			"killTask":        newTaskField(sp, newKillTaskResolver(sp), taskArgs),
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: queryType, Mutation: mutationType})
}
