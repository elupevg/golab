// Package topology provides means to model a virtual network topology.
package topology

import (
	"errors"
	"net"
	"strings"

	"github.com/goccy/go-yaml"
)

var (
	ErrUnknownNode       = errors.New("unknown node in link endpoints")
	ErrZeroNodes         = errors.New("topology must have at least one node")
	ErrTooFewEndpoints   = errors.New("link must have at least two endpoints")
	ErrInvalidEndpoint   = errors.New("invalid endpoint formatting")
	ErrInvalidIPv4Subnet = errors.New("cannot parse IPv4 subnet")
)

// Node represents a node in a virtual network topology.
type Node struct {
	Image string
}

// Link represents a link in virtual network topology.
type Link struct {
	Endpoints  []string `yaml:"endpoints"`
	IPv4Subnet string   `yaml:"ipv4_subnet"`
}

// Topology represents a virtual network consisting of nodes and links.
type Topology struct {
	Name  string          `yaml:"name"`
	Nodes map[string]Node `yaml:"nodes"`
	Links []Link          `yaml:"links"`
}

// Validate runs sanity checks to ensure that topology can be built.
func (topo Topology) Validate() error {
	if len(topo.Nodes) == 0 {
		return ErrZeroNodes
	}
	for _, link := range topo.Links {
		if len(link.Endpoints) < 2 {
			return ErrTooFewEndpoints
		}
		for _, ep := range link.Endpoints {
			ep_parts := strings.Split(ep, ":")
			if len(ep_parts) != 2 {
				return ErrInvalidEndpoint
			}
			if _, ok := topo.Nodes[ep_parts[0]]; !ok {
				return ErrUnknownNode
			}
		}
		if link.IPv4Subnet != "" {
			// Validate manually assigned IPv4 subnet
			_, _, err := net.ParseCIDR(link.IPv4Subnet)
			if err != nil {
				return ErrInvalidIPv4Subnet
			}
		}
	}
	return nil
}

// FromYAML validates and converts YAML data into a Topology struct.
func FromYAML(data []byte) (Topology, error) {
	var topo Topology
	err := yaml.Unmarshal(data, &topo)
	if err != nil {
		return Topology{}, err
	}
	if err := topo.Validate(); err != nil {
		return Topology{}, err
	}
	return topo, nil
}
