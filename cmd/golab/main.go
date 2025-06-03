package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/elupevg/golab"
	"github.com/elupevg/golab/docker"
)

const usage = "Usage:\n  golab build <topology_name>.yml\n  golab wreck <topology_name>.yml"

func main() {
	if len(os.Args) != 3 {
		fmt.Println(usage)
		os.Exit(1)
	}
	var cmd golab.Command
	switch os.Args[1] {
	case "build":
		cmd = golab.Build
	case "wreck":
		cmd = golab.Wreck
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dockerClient, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	dockerProvider := docker.New(dockerClient)
	if err := cmd(context.Background(), data, dockerProvider); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
