package docker_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/elupevg/golab/docker"
	"github.com/elupevg/golab/topology"
)

type stubDockerClient struct {
	client.APIClient
	netErr   error
	networks map[string]string
}

func (s *stubDockerClient) NetworkCreate(_ context.Context, name string, _ network.CreateOptions) (network.CreateResponse, error) {
	if s.netErr != nil {
		return network.CreateResponse{}, s.netErr
	}
	if _, ok := s.networks[name]; ok {
		return network.CreateResponse{}, fmt.Errorf("network %s already exists", name)
	}
	dummyID := strconv.Itoa(len(s.networks)+1) + "000000000000"
	s.networks[name] = dummyID
	return network.CreateResponse{ID: dummyID}, nil
}

func (s *stubDockerClient) NetworkRemove(_ context.Context, networkID string) error {
	if s.netErr != nil {
		return s.netErr
	}
	if _, ok := s.networks[networkID]; !ok {
		return fmt.Errorf("network %s does not exists", networkID)
	}
	delete(s.networks, networkID)
	return nil
}

func (s *stubDockerClient) NetworkList(_ context.Context, opts network.ListOptions) ([]network.Summary, error) {
	if s.netErr != nil {
		return nil, s.netErr
	}
	netSumms := make([]network.Summary, 0, len(s.networks))
	for name, id := range s.networks {
		netSumms = append(netSumms, network.Summary{Name: name, ID: id})
	}
	return netSumms, nil
}

func TestLinkCreateRemove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dc := &stubDockerClient{
		networks: make(map[string]string),
	}
	dp := docker.New(dc)
	link1 := topology.Link{Name: "golab-link-1", IPv4Subnet: "100.11.0.0/29", IPv4Gateway: "100.11.0.6"}
	link2 := topology.Link{Name: "golab-link-2", IPv4Subnet: "100.22.0.0/29", IPv4Gateway: "100.22.0.6"}
	// network creation
	err := dp.LinkCreate(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(dc.networks) != 1 {
		t.Fatalf("docker network count: want 1, got %d", len(dc.networks))
	}
	// network creation idempotence
	err = dp.LinkCreate(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(dc.networks) != 1 {
		t.Fatalf("docker network count: want 1, got %d", len(dc.networks))
	}
	// second network creation
	err = dp.LinkCreate(ctx, link2)
	if err != nil {
		t.Fatal(err)
	}
	if len(dc.networks) != 2 {
		t.Fatalf("docker network count: want 2, got %d", len(dc.networks))
	}
	// network removal
	err = dp.LinkRemove(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(dc.networks) != 1 {
		t.Errorf("docker network count: want 1, got %d", len(dc.networks))
	}
	// network removal idempotence
	err = dp.LinkRemove(ctx, link1)
	if err != nil {
		t.Fatal(err)
	}
	if len(dc.networks) != 1 {
		t.Errorf("docker network count: want 1, got %d", len(dc.networks))
	}
}
