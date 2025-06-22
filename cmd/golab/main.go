// Program golab orchestrates virtual network topologies based on YAML intent files.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/elupevg/golab"
	"github.com/elupevg/golab/docker"
	"github.com/elupevg/golab/logger"
)

const usage = "Usage:\n  golab build\n  golab wreck"

func main() {
	if len(os.Args) != 2 {
		fmt.Println(usage)
	}
	log := logger.New(os.Stdout, os.Stderr)
	var cmd golab.Command
	switch os.Args[1] {
	case "build":
		cmd = golab.Build
	case "wreck":
		cmd = golab.Wreck
	default:
		log.Error(fmt.Errorf("unknown command %q", os.Args[1]))
		os.Exit(1)
	}
	yamlFiles, err := filepath.Glob("*.yml")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	if len(yamlFiles) != 1 {
		log.Error(fmt.Errorf("expected one topology YAML file but found %d", len(yamlFiles)))
		os.Exit(1)
	}
	log.Success(fmt.Sprintf("found topology file %q", yamlFiles[0]))
	data, err := os.ReadFile(yamlFiles[0])
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	dockerClient, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	dockerProvider := docker.New(dockerClient, log)
	if err := cmd(context.Background(), data, dockerProvider); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
