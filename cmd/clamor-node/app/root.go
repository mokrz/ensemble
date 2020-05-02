package app

import (
	"flag"
	"fmt"
	"os"

	"github.com/mokrz/clamor/pkg/node"
)

func Execute() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	configFile := serveCmd.String("config", "/home/ubuntu/.clamor/config.json", "Path to configuration file")
	var (
		cfg        *node.Config
		cfgLoadErr error
	)

	if len(os.Args) < 2 {
		fmt.Printf("Usage: clamor [command]\n")
		return
	}

	switch os.Args[1] {
	case "serve":
		serveCmd.Parse(os.Args[2:])

		if configFile != nil {
			cfg, cfgLoadErr = node.LoadConfig(*configFile)

			if cfgLoadErr != nil {
				fmt.Printf("failed to load config with error: %s\n", cfgLoadErr.Error())
				return
			}
		}

		_ = node.New(cfg, nil).Serve()
		return
	default:
		fmt.Printf("Usage: clamor [command]\n")
		return
	}
}
