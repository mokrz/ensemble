package api

import (
	"encoding/json"
	"net/http"

	"github.com/graphql-go/graphql"
)

type Server struct {
	SockAddr string
	Schema   graphql.Schema
}

func NewServer(schema graphql.Schema, sockAddr string) (apiServer *Server) {
	return &Server{
		SockAddr: sockAddr,
		Schema:   schema,
	}
}

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
