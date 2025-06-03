// Package golab orchestrates virtual network topologies based on YAML intent files.
package golab

import (
	"context"
	"fmt"

	"github.com/elupevg/golab/topology"
)

// VirtProvider represents a virtualization provider and its methods (e.g. Docker).
type VirtProvider interface {
	LinkCreate(ctx context.Context, link topology.Link) (string, error)
	NodeCreate(ctx context.Context, node topology.Node) (string, error)
}

// Command represents a network topology orchestration command.
type Command func(ctx context.Context, data []byte, vp VirtProvider) error

// Build creates a virtual network topology described in the provided YAML intent file.
func Build(ctx context.Context, data []byte, vp VirtProvider) error {
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

// Wreck deletes a virtual network topology described in the provided YAML intent file.
func Wreck(ctx context.Context, data []byte, vp VirtProvider) error {
	return nil
}
