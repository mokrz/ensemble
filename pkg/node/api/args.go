package api

import "github.com/graphql-go/graphql"

var imageArgs = graphql.FieldConfigArgument{
	"ref": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var imagesArgs = graphql.FieldConfigArgument{
	"filter": &graphql.ArgumentConfig{
		Type:         graphql.String,
		DefaultValue: "",
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var containerArgs = graphql.FieldConfigArgument{
	"id": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var containersArgs = graphql.FieldConfigArgument{
	"filter": &graphql.ArgumentConfig{
		Type:         graphql.String,
		DefaultValue: "",
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var taskArgs = graphql.FieldConfigArgument{
	"container_id": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var tasksArgs = graphql.FieldConfigArgument{
	"filter": &graphql.ArgumentConfig{
		Type:         graphql.String,
		DefaultValue: "",
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var createImageArgs = graphql.FieldConfigArgument{
	"ref": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var createContainerArgs = graphql.FieldConfigArgument{
	"id": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"image_ref": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var createTaskArgs = graphql.FieldConfigArgument{
	"container_id": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"namespace": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}
