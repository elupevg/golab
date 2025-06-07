package docker_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/elupevg/golab/docker"
	"github.com/elupevg/golab/docker/fakeclient"
	"github.com/elupevg/golab/topology"
)

func TestLinkCreateRemove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient)
	link1 := topology.Link{
		Name: "golab-link-1",
		IPv4Subnet: &net.IPNet{
			IP:   net.ParseIP("100.11.0.0"),
			Mask: net.CIDRMask(29, 32),
		},
		IPv4Gateway: net.ParseIP("100.11.0.6"),
	}
	link2 := topology.Link{
		Name: "golab-link-2",
		IPv4Subnet: &net.IPNet{
			IP:   net.ParseIP("100.22.0.0"),
			Mask: net.CIDRMask(29, 32),
		},
		IPv4Gateway: net.ParseIP("100.22.0.6"),
	}
	// Network creation
	err := dp.LinkCreate(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 1 {
		t.Fatalf("network count: want 1, got %d", len(fakeDockerClient.Networks))
	}
	// Network creation idempotence
	err = dp.LinkCreate(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 1 {
		t.Fatalf("network count: want 1, got %d", len(fakeDockerClient.Networks))
	}
	// Second network creation
	err = dp.LinkCreate(ctx, link2)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 2 {
		t.Fatalf("network count: want 2, got %d", len(fakeDockerClient.Networks))
	}
	// Network removal
	err = dp.LinkRemove(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 1 {
		t.Fatalf("network count: want 1, got %d", len(fakeDockerClient.Networks))
	}
	// Network removal idempotence
	err = dp.LinkRemove(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 1 {
		t.Fatalf("network count: want 1, got %d", len(fakeDockerClient.Networks))
	}
}

func TestLinkExistsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient)
	wantErr := errors.New("error listing networks")
	fakeDockerClient.NetworkListErr = wantErr
	// Standalone method invocation
	_, err := dp.LinkExists(ctx, topology.Link{Name: "golab-link-1"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// Method invocation from LinkCreate
	err = dp.LinkCreate(ctx, topology.Link{Name: "golab-link-1"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
	// Method invocation from LinkRemove
	err = dp.LinkRemove(ctx, topology.Link{Name: "golab-link-1"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestLinkCreateError(t *testing.T) {
	t.Parallel()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient)
	wantErr := errors.New("failed to create a network")
	fakeDockerClient.NetworkCreateErr = wantErr
	err := dp.LinkCreate(context.Background(), topology.Link{Name: "golab-link-1"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestLinkRemoveError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient)
	err := dp.LinkCreate(ctx, topology.Link{Name: "golab-link-1"})
	if err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("failed to remove a network")
	fakeDockerClient.NetworkRemoveErr = wantErr
	err = dp.LinkRemove(ctx, topology.Link{Name: "golab-link-1"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestContainerExistsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient)
	wantErr := errors.New("error listing containers")
	fakeDockerClient.ContainerListErr = wantErr
	// Standalone method invocation
	_, err := dp.NodeExists(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	/*
		// Method invocation from LinkCreate
		err = dp.LinkCreate(ctx, topology.Link{Name: "golab-link-1"})
		if !errors.Is(err, wantErr) {
			t.Errorf("error: want %q, got %q", wantErr, err)
		}
		// Method invocation from LinkRemove
		err = dp.LinkRemove(ctx, topology.Link{Name: "golab-link-1"})
		if !errors.Is(err, wantErr) {
			t.Errorf("error: want %q, got %q", wantErr, err)
		}
	*/
}
