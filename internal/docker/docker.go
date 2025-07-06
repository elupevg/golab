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
	"github.com/elupevg/golab/internal/logger"
	"github.com/elupevg/golab/internal/topology"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// DockerProvider stores cached Docker client.
type DockerProvider struct {
	dockerClient client.APIClient
	log          *logger.Logger
}

// New returns an instance of a DockerProvider.
func New(dockerClient client.APIClient, log *logger.Logger) *DockerProvider {
	return &DockerProvider{dockerClient, log}
}

// LinkCreate translates a topology.Link entity into a Docker bridge network and creates it.
func (dp *DockerProvider) LinkCreate(ctx context.Context, link topology.Link) error {
	// Check whether network with such name already exists.
	exists, err := dp.LinkExists(ctx, link)
	if err != nil {
		return err
	}
	if exists {
		dp.log.Skipped("already created docker network " + link.Name)
		return nil
	}
	// Otherwise, create a new Docker network.
	ipamConfigs := make([]network.IPAMConfig, 0, 2)
	if link.IPv4Subnet != "" {
		ipamConfigs = append(ipamConfigs, network.IPAMConfig{
			Subnet:  link.IPv4Subnet,
			Gateway: link.IPv4Gateway,
		})
	}
	if link.IPv6Subnet != "" {
		ipamConfigs = append(ipamConfigs, network.IPAMConfig{
			Subnet:  link.IPv6Subnet,
			Gateway: link.IPv6Gateway,
		})
	}
	enableIPv4 := link.IPv4Subnet != ""
	enableIPv6 := link.IPv6Subnet != ""
	opts := network.CreateOptions{
		IPAM:       &network.IPAM{Config: ipamConfigs},
		Internal:   true, // network is internal to the Docker host.
		EnableIPv4: &enableIPv4,
		EnableIPv6: &enableIPv6,
	}
	resp, err := dp.dockerClient.NetworkCreate(ctx, link.Name, opts)
	if err != nil {
		return err
	}
	dp.log.Success(fmt.Sprintf("created docker network %s with subnets=[%v, %v], id=%s", link.Name, link.IPv4Subnet, link.IPv6Subnet, string(resp.ID[:12])))
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
		dp.log.Skipped("already removed docker network " + link.Name)
		return nil
	}
	// Otherwise, remove a Docker network.
	err = dp.dockerClient.NetworkRemove(ctx, link.Name)
	if err != nil {
		return err
	}
	dp.log.Success("removed docker network " + link.Name)
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
		ipv4Addr, _, _ := strings.Cut(iface.IPv4Addr, "/")
		ipv6Addr, _, _ := strings.Cut(iface.IPv6Addr, "/")
		endpoints[iface.Link] = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: ipv4Addr,
				IPv6Address: ipv6Addr,
			},
			DriverOpts: iface.DriverOpts,
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
		dp.log.Skipped("already created docker container " + node.Name)
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
		Sysctls:    node.Sysctls,
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
	dp.log.Success(fmt.Sprintf("started docker container %s with id=%s", node.Name, string(resp.ID[:12])))
	return nil
}

func (dp *DockerProvider) NodeRemove(ctx context.Context, node topology.Node) error {
	// Check whether container exists
	exists, err := dp.NodeExists(ctx, node)
	if err != nil {
		return err
	}
	if !exists {
		dp.log.Skipped("already removed docker container " + node.Name)
		return err
	}
	// Remove container
	err = dp.dockerClient.ContainerRemove(ctx, node.Name, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}
	dp.log.Success("removed docker container " + node.Name)
	return nil
}
