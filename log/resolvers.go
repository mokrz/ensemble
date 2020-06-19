package log

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/mokrz/clamor/node/api"
	"go.uber.org/zap"
)

// NewLoggingResolverSet handles wrapping *api.ResolverSet instances
func NewLoggingResolverSet(logger *zap.Logger, rs *api.ResolverSet) *api.ResolverSet {
	return &api.ResolverSet{
		CreateImageResolver:     NewLoggingResolver(logger, "CreateImageResolver", rs.CreateImageResolver),
		ImageResolver:           NewLoggingResolver(logger, "ImageResolver", rs.ImageResolver),
		ImagesResolver:          NewLoggingResolver(logger, "ImagesResolver", rs.ImagesResolver),
		DeleteImageResolver:     NewLoggingResolver(logger, "DeleteImageResolver", rs.DeleteImageResolver),
		CreateContainerResolver: NewLoggingResolver(logger, "CreateContainerResolver", rs.CreateContainerResolver),
		ContainerResolver:       NewLoggingResolver(logger, "ContainerResolver", rs.ContainerResolver),
		ContainersResolver:      NewLoggingResolver(logger, "ContainersResolver", rs.ContainersResolver),
		DeleteContainerResolver: NewLoggingResolver(logger, "DeleteContainerResolver", rs.DeleteContainerResolver),
		CreateTaskResolver:      NewLoggingResolver(logger, "CreateTaskResolver", rs.CreateTaskResolver),
		TaskResolver:            NewLoggingResolver(logger, "TaskResolver", rs.TaskResolver),
		TasksResolver:           NewLoggingResolver(logger, "TasksResolver", rs.TasksResolver),
		DeleteTaskResolver:      NewLoggingResolver(logger, "DeleteTaskResolver", rs.DeleteTaskResolver),
		KillTaskResolver:        NewLoggingResolver(logger, "KillTaskResolver", rs.KillTaskResolver),
	}
}

// NewLoggingResolver logs the current resolver name, its params and how long the resolver took to execute
func NewLoggingResolver(l *zap.Logger, resolver string, r graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		defer func(took time.Time) {
			var logFields []zap.Field

			for k, v := range p.Args {
				logFields = append(logFields, zap.String(k, v.(string)))
			}

			logFields = append(logFields, zap.String("took", time.Since(took).String()))
			l.Info(resolver, logFields...)
		}(time.Now())

		return r(p)
	}
}
