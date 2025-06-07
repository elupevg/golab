package fakeclient

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type Client struct {
	client.APIClient
	NetworkCreateErr   error
	NetworkRemoveErr   error
	NetworkListErr     error
	Networks           map[string]string
	ContainerCreateErr error
	ContainerRemoveErr error
	ContainerListErr   error
	Containers         map[string]string
}

func New() *Client {
	return &Client{
		Networks:   make(map[string]string, 0),
		Containers: make(map[string]string, 0),
	}
}

func (c *Client) NetworkCreate(_ context.Context, name string, _ network.CreateOptions) (network.CreateResponse, error) {
	if c.NetworkCreateErr != nil {
		return network.CreateResponse{}, c.NetworkCreateErr
	}
	if _, ok := c.Networks[name]; ok {
		return network.CreateResponse{}, fmt.Errorf("network %s already exists", name)
	}
	dummyID := strconv.Itoa(len(c.Networks)+1) + "000000000000"
	c.Networks[name] = dummyID
	return network.CreateResponse{ID: dummyID}, nil
}

func (c *Client) NetworkRemove(_ context.Context, networkID string) error {
	if c.NetworkRemoveErr != nil {
		return c.NetworkRemoveErr
	}
	if _, ok := c.Networks[networkID]; !ok {
		return fmt.Errorf("network %s does not exist", networkID)
	}
	delete(c.Networks, networkID)
	return nil
}

func (c *Client) NetworkList(_ context.Context, _ network.ListOptions) ([]network.Summary, error) {
	if c.NetworkListErr != nil {
		return nil, c.NetworkListErr
	}
	netSumms := make([]network.Summary, 0, len(c.Networks))
	for name, id := range c.Networks {
		netSumms = append(netSumms, network.Summary{Name: name, ID: id})
	}
	return netSumms, nil
}

func (c *Client) ContainerCreate(_ context.Context, _ *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *ocispec.Platform, name string) (container.CreateResponse, error) {
	if c.ContainerCreateErr != nil {
		return container.CreateResponse{}, c.ContainerCreateErr
	}
	if _, ok := c.Containers[name]; ok {
		return container.CreateResponse{}, fmt.Errorf("container %s already exists", name)
	}
	dummyID := strconv.Itoa(len(c.Containers)+1) + "000000000000"
	c.Containers[name] = dummyID
	return container.CreateResponse{ID: dummyID}, nil
}

func (c *Client) ContainerRemove(_ context.Context, containerID string, _ container.RemoveOptions) error {
	if c.ContainerRemoveErr != nil {
		return c.ContainerRemoveErr
	}
	if _, ok := c.Containers[containerID]; !ok {
		return fmt.Errorf("container %s does not exist", containerID)
	}
	delete(c.Containers, containerID)
	return nil
}

func (c *Client) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	if c.ContainerListErr != nil {
		return nil, c.ContainerListErr
	}
	contSumms := make([]container.Summary, 0, len(c.Containers))
	for name, id := range c.Containers {
		contSumms = append(contSumms, container.Summary{Names: []string{"/" + name}, ID: id})
	}
	return contSumms, nil
}
