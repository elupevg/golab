package docker_test

import (
	"context"
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
	netCount int
}

func (s *stubDockerClient) NetworkCreate(_ context.Context, name string, _ network.CreateOptions) (network.CreateResponse, error) {
	if s.netErr != nil {
		return network.CreateResponse{}, s.netErr
	}
	s.netCount++
	return network.CreateResponse{ID: strconv.Itoa(s.netCount) + "000000000000"}, nil
}

func (s *stubDockerClient) NetworkRemove(_ context.Context, networkID string) error {
	if s.netErr != nil {
		return s.netErr
	}
	s.netCount--
	return nil
}

func TestLinkCreateRemove(t *testing.T) {
	t.Parallel()
	dc := new(stubDockerClient)
	dp := docker.New(dc)
	link1 := topology.Link{
		Name:        "golab-link-1",
		IPv4Subnet:  "100.11.0.0/29",
		IPv4Gateway: "100.11.0.6",
	}
	link2 := topology.Link{
		Name:        "golab-link-2",
		IPv4Subnet:  "100.22.0.0/29",
		IPv4Gateway: "100.22.0.6",
	}
	err := dp.LinkCreate(context.Background(), link1)
	if err != nil {
		t.Fatal(err)
	}
	err = dp.LinkCreate(context.Background(), link2)
	if err != nil {
		t.Fatal(err)
	}
	if dc.netCount != 2 {
		t.Fatalf("docker network count: want 2, got %d", dc.netCount)
	}
	err = dp.LinkRemove(context.Background(), link1)
	if err != nil {
		t.Fatal(err)
	}
	if dc.netCount != 1 {
		t.Errorf("docker network count: want 1, got %d", dc.netCount)
	}
}
