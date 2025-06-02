// Package golab orchestrates virtual network topologies based on YAML intent files.
package golab

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/elupevg/golab/docker"
	"github.com/elupevg/golab/topology"
)

const usage = "Usage:\n  golab build lab.yml\n  golab wreck lab.yml"

// virtProvider represents a virtualization provider and its methods (e.g. Docker).
type virtProvider interface {
	LinkCreate(ctx context.Context, link topology.Link) (string, error)
	NodeCreate(ctx context.Context, node topology.Node) (string, error)
}

// Build creates a virtual network topology described in provided YAML intent file.
func Build(ctx context.Context, data []byte, vp virtProvider) error {
	topo, err := topology.FromYAML(data)
	if err != nil {
		return err
	}
	for _, link := range topo.Links {
		id, err := vp.LinkCreate(ctx, link)
		if err != nil {
			return err
		}
		fmt.Println("Successfully created link", id)
	}
	for _, node := range topo.Nodes {
		id, err := vp.NodeCreate(ctx, node)
		if err != nil {
			return err
		}
		fmt.Println("Successfully created node", id)
	}
	return nil
}

// Wreck deletes a virtual network topology described in provided YAML intent file.
func Wreck(ctx context.Context, data []byte, vp virtProvider) error {
	return nil
}

type action func(ctx context.Context, data []byte, vp virtProvider) error

func Main() int {
	if len(os.Args) != 3 {
		fmt.Println(usage)
		return 1
	}
	var cmd action
	switch os.Args[1] {
	case "build":
		cmd = Build
	case "wreck":
		cmd = Wreck
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.47"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	vp := docker.New(dockerClient)
	if err := cmd(context.Background(), data, vp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return 0
}
