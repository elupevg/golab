package docker_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/elupevg/golab/internal/docker"
	"github.com/elupevg/golab/internal/docker/fakeclient"
	"github.com/elupevg/golab/internal/logger"
	"github.com/elupevg/golab/internal/topology"
)

func TestLinkCreateRemove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	link := topology.Link{
		Name: "golab-link-1",
		Subnets: []*net.IPNet{
			{
				IP:   net.ParseIP("100.11.0.0"),
				Mask: net.CIDRMask(29, 32),
			},
			{
				IP:   net.ParseIP("2001:db8:11::"),
				Mask: net.CIDRMask(64, 128),
			},
		},
		Gateways: []net.IP{
			net.ParseIP("100.11.0.6"),
			net.ParseIP("2001:db8:11::ffff:ffff:ffff:fffe"),
		},
	}
	// link creation
	err := dp.LinkCreate(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	exists, err := dp.LinkExists(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("link does not exist after creation")
	}
	// link creation idempotence
	err = dp.LinkCreate(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 1 {
		t.Fatalf("network count: want 1, got %d", len(fakeDockerClient.Networks))
	}
	// link removal
	err = dp.LinkRemove(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	exists, err = dp.LinkExists(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("link exist after removal")
	}
	// link removal idempotence
	err = dp.LinkRemove(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 0 {
		t.Errorf("network count: want 0, got %d", len(fakeDockerClient.Networks))
	}
}

func TestLinkExistsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	wantErr := errors.New("failed to list networks")
	fakeDockerClient.NetworkListErr = wantErr
	_, err := dp.LinkExists(ctx, topology.Link{Name: "golab-link-1"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestLinkCreateError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	link := topology.Link{
		Subnets: []*net.IPNet{
			{
				IP:   net.ParseIP("100.11.0.0"),
				Mask: net.CIDRMask(29, 32),
			},
		},
		Gateways: []net.IP{net.ParseIP("100.11.0.6")},
	}
	// network list error
	wantErr := errors.New("failed to list networks")
	fakeDockerClient.NetworkListErr = wantErr
	err := dp.LinkCreate(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// network create error
	wantErr = errors.New("failed to create a network")
	fakeDockerClient.NetworkListErr = nil
	fakeDockerClient.NetworkCreateErr = wantErr
	err = dp.LinkCreate(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestLinkRemoveErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	link := topology.Link{
		Name: "golab-link-1",
		Subnets: []*net.IPNet{
			{
				IP:   net.ParseIP("100.11.0.0"),
				Mask: net.CIDRMask(29, 32),
			},
		},
		Gateways: []net.IP{net.ParseIP("100.11.0.6")},
	}
	// create test network
	err := dp.LinkCreate(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	// network list error
	wantErr := errors.New("failed to list networks")
	fakeDockerClient.NetworkListErr = wantErr
	err = dp.LinkRemove(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// network remove error
	wantErr = errors.New("failed to remove a network")
	fakeDockerClient.NetworkListErr = nil
	fakeDockerClient.NetworkRemoveErr = wantErr
	err = dp.LinkRemove(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestNodeCreateRemove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	node := topology.Node{
		Name:  "frr01",
		Image: "quay.io/frrouting/frr:master",
		Binds: []string{"/lib/modules:/lib/modules"},
		Interfaces: []*topology.Interface{
			{
				Name: "lo",
				IPv4: "192.168.0.1/32",
				IPv6: "2001:db8:192:168::1/128",
			},
			{
				Name: "eth0",
				Link: "golab-link-1",
				IPv4: "100.64.0.1/29",
				IPv6: "2001:db8:64::1/64",
			},
		},
	}
	// node creation
	err := dp.NodeCreate(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	exists, err := dp.NodeExists(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("node does not exist after creation")
	}
	// node creation idempotence
	err = dp.NodeCreate(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Containers) != 1 {
		t.Fatalf("container count: want 1, got %d", len(fakeDockerClient.Containers))
	}
	// node removal
	err = dp.NodeRemove(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	exists, err = dp.NodeExists(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("node exists after removal")
	}
	// node removal idempotence
	err = dp.NodeRemove(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	if len(fakeDockerClient.Networks) != 0 {
		t.Errorf("container count: want 0, got %d", len(fakeDockerClient.Containers))
	}
}

func TestNodeExistsError(t *testing.T) {
	t.Parallel()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	wantErr := errors.New("failed to list containers")
	fakeDockerClient.ContainerListErr = wantErr
	_, err := dp.NodeExists(context.Background(), topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestNodeCreateError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	// container list error
	wantErr := errors.New("failed to list containers")
	fakeDockerClient.ContainerListErr = wantErr
	err := dp.NodeCreate(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// container create error
	wantErr = errors.New("failed to create a container")
	fakeDockerClient.ContainerListErr = nil
	fakeDockerClient.ContainerCreateErr = wantErr
	err = dp.NodeCreate(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// container start error
	wantErr = errors.New("failed to start a container")
	fakeDockerClient.ContainerCreateErr = nil
	fakeDockerClient.ContainerStartErr = wantErr
	err = dp.NodeCreate(context.Background(), topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestNodeRemoveError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fakeDockerClient := fakeclient.New()
	dp := docker.New(fakeDockerClient, logger.New(io.Discard, io.Discard))
	// create test container
	err := dp.NodeCreate(ctx, topology.Node{Name: "frr01"})
	if err != nil {
		t.Fatal(err)
	}
	// container list error
	wantErr := errors.New("failed to list containers")
	fakeDockerClient.ContainerListErr = wantErr
	err = dp.NodeRemove(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// container remove error
	wantErr = errors.New("failed to remove a container")
	fakeDockerClient.ContainerListErr = nil
	fakeDockerClient.ContainerRemoveErr = wantErr
	err = dp.NodeRemove(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}
