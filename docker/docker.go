// Package docker translates GoLab network topology entities into Docker objects.
// Examples:
//
//	topology.Link is equivalent to a Docker bridge network
//	topology.Node is equivalent to a Docker container
package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/elupevg/golab/topology"
)

// DockerProvider stores cached Docker client.
type DockerProvider struct {
	dockerClient client.APIClient
}

// New returns an instance of a DockerProvider.
func New(dockerClient client.APIClient) *DockerProvider {
	return &DockerProvider{dockerClient: dockerClient}
}

// LinkCreate translates a topology.Link entity into a Docker bridge network and creates it.
func (dp *DockerProvider) LinkCreate(ctx context.Context, link topology.Link) error {
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
		return err
	}
	fmt.Printf("Created Docker network: name=%s, subnet=%s, id=%s\n", link.Name, link.IPv4Subnet, string(resp.ID[:12]))
	if resp.Warning != "" {
		fmt.Println("Warnings: ", resp.Warning)
	}
	return nil
}

// LinkRemove translates a topology.Link entity into a Docker bridge network and removes it.
func (dp *DockerProvider) LinkRemove(ctx context.Context, link topology.Link) error {
	err := dp.dockerClient.NetworkRemove(ctx, link.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Removed Docker network: name=%s\n", link.Name)
	return nil
}

// NodeCreate translates a topology.Node entity into a Docker container and creates/starts it.
func (dp *DockerProvider) NodeCreate(_ context.Context, _ topology.Node) error {
	return nil
}
