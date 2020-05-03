package app

import (
	"flag"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/mokrz/clamor/pkg/node"
)

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

	if serveErr := node.NewNode(cfg, ctr).Serve(); serveErr != nil {
		fmt.Printf("node.Serve() failed with error: %s\n", serveErr.Error())
		return
	}

	return
}
