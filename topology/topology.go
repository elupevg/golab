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
	ErrZeroNodes         = errors.New("topology has no nodes defined")
	ErrTooFewEndpoints   = errors.New("link has less than two endpoints")
	ErrInvalidEndpoint   = errors.New("invalid endpoint format")
	ErrInvalidIPv4Subnet = errors.New("cannot parse IPv4 subnet")
)

// Node represents a node in a virtual network topology.
type Node struct {
	Image string
}

// Link represents a link in a virtual network topology.
type Link struct {
	Endpoints   []string `yaml:"endpoints"`
	Name        string   `yaml:"name"`
	IPv4Subnet  string   `yaml:"ipv4_subnet"`
	IPv4Gateway string   `yaml:"ipv4_gateway"`
}

// Topology represents a virtual network comprised of nodes and links.
type Topology struct {
	Name  string          `yaml:"name"`
	Nodes map[string]Node `yaml:"nodes"`
	Links []Link          `yaml:"links"`
}

// validate runs sanity checks to ensure that the network topology can be built.
// It is a value method since validation should not modify the original struct.
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
// Returning a struct value instead of a pointer is intentional as
// topology is not supposed to be modified by the caller.
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
