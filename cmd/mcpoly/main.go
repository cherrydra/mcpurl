package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/cherrydra/mcpurl/cmd/mcpoly/config"
	"github.com/cherrydra/mcpurl/cmd/mcpoly/server"
)

var configFile string

func main() {
	flag.StringVar(&configFile, "f", "mcp.json", "Path to the mcp servers config file")
	flag.Parse()

	conf, err := config.Parse(configFile)
	if err != nil {
		exit(err)
	}

	rp := &server.ReverseProxy{}

	ctx := context.TODO()

	if err := rp.AddBackends(ctx, conf.Servers); err != nil {
		exit(err)
	}

	if err := rp.Run(ctx); err != nil {
		exit(err)
	}
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
