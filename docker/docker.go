// Package docker translates GoLab network topology entities into Docker objects.
// Examples:
//
//	topology.Link is equivalent to a Docker bridge network
//	topology.Node is equivalent to a Docker container
package docker

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/elupevg/golab/topology"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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
	// Check whether network with such name already exists.
	exists, err := dp.LinkExists(ctx, link)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("[SKIPPED] Docker network %q already exists\n", link.Name)
		return nil
	}
	// Otherwise, create a new Docker network.
	opts := network.CreateOptions{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  link.IPv4Subnet.String(),
					Gateway: link.IPv4Gateway.String(),
				},
			},
		},
	}
	resp, err := dp.dockerClient.NetworkCreate(ctx, link.Name, opts)
	if err != nil {
		return err
	}
	fmt.Printf("[SUCCESS] created Docker network %q: subnet=%s, id=%s\n", link.Name, link.IPv4Subnet, string(resp.ID[:12]))
	return nil
}

// LinkExists checks whether a Docker network representing the provided topology.Link already exists.
func (dp *DockerProvider) LinkExists(ctx context.Context, link topology.Link) (bool, error) {
	netSums, err := dp.dockerClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, netSum := range netSums {
		if netSum.Name == link.Name {
			return true, nil
		}
	}
	return false, nil
}

// LinkRemove translates a topology.Link entity into a Docker bridge network and removes it.
func (dp *DockerProvider) LinkRemove(ctx context.Context, link topology.Link) error {
	// Check whether network with such name exists.
	exists, err := dp.LinkExists(ctx, link)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Printf("[SKIPPED] Docker network %q already removed\n", link.Name)
		return nil
	}
	// Otherwise, remove a Docker network.
	err = dp.dockerClient.NetworkRemove(ctx, link.Name)
	if err != nil {
		return err
	}
	fmt.Printf("[SUCCESS] removed Docker network %q\n", link.Name)
	return nil
}

// NodeExists checks whether a Docker container representing the provided topology.Node already exists.
func (dp *DockerProvider) NodeExists(ctx context.Context, node topology.Node) (bool, error) {
	contSums, err := dp.dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, err
	}
	for _, contSum := range contSums {
		if slices.Contains(contSum.Names, "/"+node.Name) {
			return true, nil
		}
	}
	return false, nil
}

// generateMounts converts list of binds from YAML topology file into a slice of Docker mounts.
func generateMounts(node topology.Node) []mount.Mount {
	mounts := make([]mount.Mount, 0, len(node.Binds))
	for _, bind := range node.Binds {
		parts := strings.Split(bind, ":")
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: parts[0],
			Target: parts[1],
		})
	}
	return mounts
}

// generateNetworkConfig converts node configuration into Docker container network configuration.
func generateNetworkConfig(node topology.Node) *network.NetworkingConfig {
	endpoints := make(map[string]*network.EndpointSettings, len(node.Interfaces))
	for _, iface := range node.Interfaces {
		endpoints[iface.Link] = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: iface.IPv4.String(),
			},
		}
	}
	return &network.NetworkingConfig{EndpointsConfig: endpoints}
}

// NodeCreate translates a topology.Node entity into a Docker container and creates/starts it.
func (dp *DockerProvider) NodeCreate(ctx context.Context, node topology.Node) error {
	// Check if container already exists
	exists, err := dp.NodeExists(ctx, node)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("[SKIPPED] Docker container %q already exists\n", node.Name)
		return nil
	}
	// Generate new container configuration
	contConfig := &container.Config{
		Hostname: node.Name,
		Image:    node.Image,
	}
	initialize := true
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Privileged: true,
		Init:       &initialize,
		Mounts:     generateMounts(node),
	}
	netConfig := generateNetworkConfig(node)
	platform := new(ocispec.Platform)
	// Create new container
	resp, err := dp.dockerClient.ContainerCreate(ctx, contConfig, hostConfig, netConfig, platform, node.Name)
	if err != nil {
		return err
	}
	// Start new container
	err = dp.dockerClient.ContainerStart(ctx, node.Name, container.StartOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("[SUCCESS] started Docker container %q: ID=%s\n", node.Name, string(resp.ID[:12]))
	return nil
}

func (dp *DockerProvider) NodeRemove(ctx context.Context, node topology.Node) error {
	// Check whether container exists
	exists, err := dp.NodeExists(ctx, node)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Printf("[SKIPPED] Docker container %q already removed\n", node.Name)
		return err
	}
	// Remove container
	err = dp.dockerClient.ContainerRemove(ctx, node.Name, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}
	fmt.Printf("[SUCCESS] removed Docker container %q\n", node.Name)
	return nil
}
