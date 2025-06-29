package docker_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/elupevg/golab/internal/docker"
	"github.com/elupevg/golab/internal/logger"
	"github.com/elupevg/golab/internal/topology"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type fakeDockerClient struct {
	client.APIClient
	networkCreateErr   error
	networkRemoveErr   error
	networkListErr     error
	networks           map[string]string
	containerCreateErr error
	containerStartErr  error
	containerRemoveErr error
	containerListErr   error
	containers         map[string]string
}

func newFakeDockerClient() *fakeDockerClient {
	return &fakeDockerClient{
		networks:   make(map[string]string, 0),
		containers: make(map[string]string, 0),
	}
}

func (f *fakeDockerClient) NetworkCreate(_ context.Context, name string, _ network.CreateOptions) (network.CreateResponse, error) {
	if f.networkCreateErr != nil {
		return network.CreateResponse{}, f.networkCreateErr
	}
	if _, ok := f.networks[name]; ok {
		return network.CreateResponse{}, fmt.Errorf("network %s already exists", name)
	}
	dummyID := strconv.Itoa(len(f.networks)+1) + "000000000000"
	f.networks[name] = dummyID
	return network.CreateResponse{ID: dummyID}, nil
}

func (f *fakeDockerClient) NetworkRemove(_ context.Context, networkID string) error {
	if f.networkRemoveErr != nil {
		return f.networkRemoveErr
	}
	if _, ok := f.networks[networkID]; !ok {
		return fmt.Errorf("network %s does not exist", networkID)
	}
	delete(f.networks, networkID)
	return nil
}

func (f *fakeDockerClient) NetworkList(_ context.Context, _ network.ListOptions) ([]network.Summary, error) {
	if f.networkListErr != nil {
		return nil, f.networkListErr
	}
	netSumms := make([]network.Summary, 0, len(f.networks))
	for name, id := range f.networks {
		netSumms = append(netSumms, network.Summary{Name: name, ID: id})
	}
	return netSumms, nil
}

func (f *fakeDockerClient) ContainerCreate(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, name string) (container.CreateResponse, error) {
	if f.containerCreateErr != nil {
		return container.CreateResponse{}, f.containerCreateErr
	}
	if _, ok := f.containers[name]; ok {
		return container.CreateResponse{}, fmt.Errorf("container %s already exists", name)
	}
	dummyID := strconv.Itoa(len(f.containers)+1) + "000000000000"
	f.containers[name] = dummyID
	return container.CreateResponse{ID: dummyID}, nil
}

func (f *fakeDockerClient) ContainerStart(_ context.Context, containerID string, _ container.StartOptions) error {
	if f.containerStartErr != nil {
		return f.containerStartErr
	}
	if _, ok := f.containers[containerID]; !ok {
		return fmt.Errorf("container %s does not exists", containerID)
	}
	return nil
}

func (f *fakeDockerClient) ContainerRemove(_ context.Context, containerID string, _ container.RemoveOptions) error {
	if f.containerRemoveErr != nil {
		return f.containerRemoveErr
	}
	if _, ok := f.containers[containerID]; !ok {
		return fmt.Errorf("container %s does not exist", containerID)
	}
	delete(f.containers, containerID)
	return nil
}

func (f *fakeDockerClient) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	if f.containerListErr != nil {
		return nil, f.containerListErr
	}
	contSumms := make([]container.Summary, 0, len(f.containers))
	for name, id := range f.containers {
		contSumms = append(contSumms, container.Summary{Names: []string{"/" + name}, ID: id})
	}
	return contSumms, nil
}

func TestLinkCreateRemove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
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
	if len(fdc.networks) != 1 {
		t.Fatalf("network count: want 1, got %d", len(fdc.networks))
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
	if len(fdc.networks) != 0 {
		t.Errorf("network count: want 0, got %d", len(fdc.networks))
	}
}

func TestLinkExistsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
	wantErr := errors.New("failed to list networks")
	fdc.networkListErr = wantErr
	_, err := dp.LinkExists(ctx, topology.Link{Name: "golab-link-1"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestLinkCreateError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
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
	fdc.networkListErr = wantErr
	err := dp.LinkCreate(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// network create error
	wantErr = errors.New("failed to create a network")
	fdc.networkListErr = nil
	fdc.networkCreateErr = wantErr
	err = dp.LinkCreate(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestLinkRemoveErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
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
	fdc.networkListErr = wantErr
	err = dp.LinkRemove(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// network remove error
	wantErr = errors.New("failed to remove a network")
	fdc.networkListErr = nil
	fdc.networkRemoveErr = wantErr
	err = dp.LinkRemove(ctx, link)
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestNodeCreateRemove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
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
	if len(fdc.containers) != 1 {
		t.Fatalf("container count: want 1, got %d", len(fdc.containers))
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
	if len(fdc.networks) != 0 {
		t.Errorf("container count: want 0, got %d", len(fdc.containers))
	}
}

func TestNodeExistsError(t *testing.T) {
	t.Parallel()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
	wantErr := errors.New("failed to list containers")
	fdc.containerListErr = wantErr
	_, err := dp.NodeExists(context.Background(), topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestNodeCreateError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
	// container list error
	wantErr := errors.New("failed to list containers")
	fdc.containerListErr = wantErr
	err := dp.NodeCreate(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// container create error
	wantErr = errors.New("failed to create a container")
	fdc.containerListErr = nil
	fdc.containerCreateErr = wantErr
	err = dp.NodeCreate(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// container start error
	wantErr = errors.New("failed to start a container")
	fdc.containerCreateErr = nil
	fdc.containerStartErr = wantErr
	err = dp.NodeCreate(context.Background(), topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestNodeRemoveError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fdc := newFakeDockerClient()
	dp := docker.New(fdc, logger.New(io.Discard, io.Discard))
	// create test container
	err := dp.NodeCreate(ctx, topology.Node{Name: "frr01"})
	if err != nil {
		t.Fatal(err)
	}
	// container list error
	wantErr := errors.New("failed to list containers")
	fdc.containerListErr = wantErr
	err = dp.NodeRemove(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
	// container remove error
	wantErr = errors.New("failed to remove a container")
	fdc.containerListErr = nil
	fdc.containerRemoveErr = wantErr
	err = dp.NodeRemove(ctx, topology.Node{Name: "frr01"})
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}
