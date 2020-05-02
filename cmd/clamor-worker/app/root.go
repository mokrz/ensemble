package app

import(
	"flag"
	"fmt"
	"os"
	"github.com/mokrz/clamor/pkg/node"
)

func Execute() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	_ = serveCmd.String("config", "/etc/clamor/config.yaml", "Path to configuration file")

	if len(os.Args) < 2 {
		fmt.Printf("Usage: clamor [command]\n")
        return
	}

	switch os.Args[1] {
	case "serve":
		serveCmd.Parse(os.Args[2:])
		_ = node.New().Serve()
		return
	default:
		fmt.Printf("Usage: clamor [command]\n")
        return
	}
}