package api

import (
	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/node"
)

// NewGraphQLSchema returns a new graphql schema instance containing root Query and root Mutation types.
// It's responsible for allocating the remainder of the API's graphql fields and wiring them to their respective resolvers + arguments.
func NewGraphQLSchema(ns node.Service, resolverSet *ResolverSet) (schema graphql.Schema, err error) {
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"image":      NewImageField(ns, resolverSet.ImageResolver, imageArgs),
			"images":     NewImagesField(ns, resolverSet.ImagesResolver, imagesArgs),
			"container":  NewContainerField(ns, resolverSet.ContainerResolver, containerArgs),
			"containers": NewContainersField(ns, resolverSet.ContainersResolver, containersArgs),
			"task":       NewTaskField(ns, resolverSet.TaskResolver, taskArgs),
			"tasks":      NewTasksField(ns, resolverSet.TasksResolver, tasksArgs),
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createImage":     NewImageField(ns, resolverSet.CreateImageResolver, createImageArgs),
			"createContainer": NewContainerField(ns, resolverSet.ContainerResolver, createContainerArgs),
			"createTask":      NewTaskField(ns, resolverSet.CreateTaskResolver, createTaskArgs),
			"deleteImage":     NewImageField(ns, resolverSet.DeleteImageResolver, imageArgs),
			"deleteContainer": NewContainerField(ns, resolverSet.DeleteContainerResolver, containerArgs),
			"deleteTask":      NewTaskField(ns, resolverSet.DeleteTaskResolver, taskArgs),
			"killTask":        NewTaskField(ns, resolverSet.KillTaskResolver, taskArgs),
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: queryType, Mutation: mutationType})
}
