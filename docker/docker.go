package docker

import (
	"context"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/elupevg/golab/topology"
)

type DockerProvider struct {
	dockerClient client.APIClient
}

func New(dockerClient client.APIClient) *DockerProvider {
	return &DockerProvider{dockerClient: dockerClient}
}

func (dp *DockerProvider) LinkCreate(ctx context.Context, link topology.Link) (string, error) {
	opts := network.CreateOptions{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  link.IPv4Subnet,
					Gateway: link.IPv4Gateway,
				},
			},
		},
	}
	resp, err := dp.dockerClient.NetworkCreate(ctx, link.Name, opts)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (dp *DockerProvider) NodeCreate(_ context.Context, _ topology.Node) (string, error) {
	return "", nil
}
