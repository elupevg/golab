package fakeclient

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Client struct {
	client.APIClient
	NetworkCreateErr error
	NetworkRemoveErr error
	NetworkListErr   error
	Networks         map[string]string
}

func New() *Client {
	return &Client{Networks: make(map[string]string, 0)}
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
		return fmt.Errorf("network %s does not exists", networkID)
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
