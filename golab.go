// Package golab orchestrates virtual network topologies based on YAML intent files.
package golab

import (
	"context"

	"github.com/elupevg/golab/topology"
)

// VirtProvider represents a virtualization provider and its methods (e.g. Docker).
type VirtProvider interface {
	LinkCreate(ctx context.Context, link topology.Link) error
	LinkRemove(ctx context.Context, link topology.Link) error
	NodeCreate(ctx context.Context, node topology.Node) error
	NodeRemove(ctx context.Context, node topology.Node) error
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
		err := vp.LinkCreate(ctx, *link)
		if err != nil {
			return err
		}
	}
	for _, node := range topo.Nodes {
		err := vp.NodeCreate(ctx, *node)
		if err != nil {
			return err
		}
	}
	return nil
}

// Wreck deletes a virtual network topology described in the provided YAML intent file.
func Wreck(ctx context.Context, data []byte, vp VirtProvider) error {
	topo, err := topology.FromYAML(data)
	if err != nil {
		return err
	}
	for _, node := range topo.Nodes {
		err := vp.NodeRemove(ctx, *node)
		if err != nil {
			return err
		}
	}
	for _, link := range topo.Links {
		err := vp.LinkRemove(ctx, *link)
		if err != nil {
			return err
		}
	}
	return nil
}
