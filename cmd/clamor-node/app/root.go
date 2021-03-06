package app

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/containerd/containerd"
	"github.com/mokrz/clamor/log"
	"github.com/mokrz/clamor/node"
	"github.com/mokrz/clamor/node/api"
	node_api "github.com/mokrz/clamor/node/api"
	"go.uber.org/zap"
)

// Execute runs the root clamor-node logic.
// It's responsible for loading the given configuration, allocating dependencies and starting the clamor-node API daemon.
func Execute() {
	configFile := flag.String("config", "/home/ubuntu/.clamor/config.json", "Path to configuration file")
	var (
		cfg        *node.Config
		cfgLoadErr error
	)

	flag.Parse()

	if configFile != nil {
		cfg, cfgLoadErr = node.LoadConfig(*configFile)

		if cfgLoadErr != nil {
			fmt.Printf("node.LoadConfig failed with error: %s\n", cfgLoadErr.Error())
			return
		}
	}

	ctr, ctrErr := containerd.New("/run/containerd/containerd.sock")

	if ctrErr != nil {
		fmt.Printf("containerd.New failed with error: %s\n", ctrErr.Error())
		return
	}

	defer ctr.Close()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	nodeSvc := node.NewNode(ctr)
	nodeSvc = log.NewLoggingNode(logger, nodeSvc)

	resolverSet := api.NewResolverSet(nodeSvc)
	resolverSet = log.NewLoggingResolverSet(logger, resolverSet)

	gqlSchema, gqlSchemaErr := node_api.NewGraphQLSchema(nodeSvc, resolverSet)

	if gqlSchemaErr != nil {
		fmt.Printf("node_api.NewGraphQLSchema failed with error: %s\n", gqlSchemaErr.Error())
		return
	}

	apiServer := node_api.NewServer(gqlSchema, cfg.APIHost+":"+strconv.Itoa(cfg.APIPort))

	if serveErr := apiServer.Serve(); serveErr != nil {
		fmt.Printf("node.Serve() failed with error: %s\n", serveErr.Error())
		return
	}

	return
}
