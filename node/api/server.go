/*
Package api defines Node GraphQL fields, their resolvers, their arguments and their exposure.
*/
package api

import (
	"encoding/json"
	"net/http"

	"github.com/graphql-go/graphql"
)

// Server holds various API server resources.
type Server struct {
	SockAddr string
	Schema   graphql.Schema
}

// NewServer returns Server instances.
func NewServer(schema graphql.Schema, sockAddr string) (apiServer *Server) {
	return &Server{
		SockAddr: sockAddr,
		Schema:   schema,
	}
}

// Serve graphql requests over HTTP on the Server instance's SockAddr.
func (as Server) Serve() (err error) {
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := graphql.Do(graphql.Params{
			Schema:        as.Schema,
			RequestString: r.URL.Query().Get("query"),
		})

		json.NewEncoder(w).Encode(result)
	})

	return http.ListenAndServe(as.SockAddr, nil)
}
