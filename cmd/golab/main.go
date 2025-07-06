// Program golab orchestrates virtual network topologies based on YAML intent files.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/elupevg/golab/configen"
	"github.com/elupevg/golab/docker"
	"github.com/elupevg/golab/logger"
	"github.com/elupevg/golab/orchestrator"
)

const usage = "Usage:\n  golab build\n  golab wreck"

func main() {
	log := logger.New(os.Stdout, os.Stderr)
	if len(os.Args) != 2 {
		fmt.Println(usage)
		return
	}
	var cmd orchestrator.Command
	switch os.Args[1] {
	case "build":
		cmd = orchestrator.Build
	case "wreck":
		cmd = orchestrator.Wreck
	default:
		log.Errored(fmt.Errorf("unknown command %q", os.Args[1]))
		os.Exit(1)
	}
	yamlFiles, err := filepath.Glob("*.yml")
	if err != nil {
		log.Errored(err)
		os.Exit(1)
	}
	if len(yamlFiles) != 1 {
		log.Errored(fmt.Errorf("expected 1 topology YAML file but found %d", len(yamlFiles)))
		os.Exit(1)
	}
	log.Success(fmt.Sprintf("found topology file %s", yamlFiles[0]))
	data, err := os.ReadFile(yamlFiles[0])
	if err != nil {
		log.Errored(err)
		os.Exit(1)
	}
	dockerClient, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Errored(err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	dockerProvider := docker.New(dockerClient, log)
	configProvider := configen.New(log)
	if err := cmd(context.Background(), data, dockerProvider, configProvider); err != nil {
		log.Errored(err)
		os.Exit(1)
	}
}
