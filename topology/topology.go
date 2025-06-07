// Package topology provides means to model a virtual network topology.
package topology

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/goccy/go-yaml"
)

const linkNamePrefix = "golab-link"

var (
	ErrCorruptYAML      = errors.New("cannot parse YAML file")
	ErrUnknownNode      = errors.New("unknown node in a link endpoint")
	ErrZeroNodes        = errors.New("topology has no nodes defined")
	ErrTooFewEndpoints  = errors.New("link has less than two endpoints")
	ErrInvalidEndpoint  = errors.New("invalid endpoint format")
	ErrInvalidIP        = errors.New("cannot parse IP address/prefix")
	ErrInvalidGW        = errors.New("invalid gateway IP address")
	ErrInvalidInterface = errors.New("invalid interface name")
)

// Node represents a node in a virtual network topology.
type Node struct {
	Name  string
	Image string `yaml:"image"`
	Links []string
	Binds []string `yaml:"binds"`
}

// Link represents a link in a virtual network topology.
type Link struct {
	Endpoints   []string `yaml:"endpoints"`
	Name        string   `yaml:"name"`
	IPv4Subnet  string   `yaml:"ipv4_subnet"`
	IPv4Gateway string   `yaml:"ipv4_gateway"`
}

// validateIP verifies the sanity of the link IP configuration.
func (l Link) validateIP() error {
	ipv4net, err := netip.ParsePrefix(l.IPv4Subnet)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidIP, l.IPv4Subnet)
	}
	ipv4gw, err := netip.ParseAddr(l.IPv4Gateway)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidIP, l.IPv4Gateway)
	}
	if !ipv4net.Contains(ipv4gw) {
		return fmt.Errorf("%w: %s not in %s", ErrInvalidGW, l.IPv4Gateway, l.IPv4Subnet)
	}
	return nil
}

// Topology represents a virtual network comprised of nodes and links.
type Topology struct {
	Name  string           `yaml:"name"`
	Nodes map[string]*Node `yaml:"nodes"`
	Links []*Link          `yaml:"links"`
}

// validator represents a topology sanity check.
type validator func() error

// validateNodes checks whether topology defines at least one node.
func (topo Topology) validateNodes() error {
	if len(topo.Nodes) == 0 {
		return ErrZeroNodes
	}
	return nil
}

// validateEndpoint checks whether the provided string represents a valid network endpoint.
func (topo Topology) validateEndpoint(ep string) error {
	parts := strings.Split(ep, ":")
	if len(parts) != 2 {
		return fmt.Errorf("%w: %q", ErrInvalidEndpoint, ep)
	}
	node, iface := parts[0], parts[1]
	if _, ok := topo.Nodes[node]; !ok {
		return fmt.Errorf("%w: %q in %q", ErrUnknownNode, node, ep)
	}
	if !strings.HasPrefix(iface, "eth") {
		return fmt.Errorf("%w: %q in %q", ErrInvalidInterface, iface, ep)
	}
	return nil
}

// validateLinks checks endpoints and IP data associated with each link.
func (topo Topology) validateLinks() error {
	for _, link := range topo.Links {
		if len(link.Endpoints) < 2 {
			return ErrTooFewEndpoints
		}
		for _, ep := range link.Endpoints {
			if err := topo.validateEndpoint(ep); err != nil {
				return err
			}
		}
		if err := link.validateIP(); err != nil {
			return err
		}
	}
	return nil
}

// validate runs sanity checks to ensure that the network topology can be built.
func (topo Topology) validate() error {
	validators := []validator{
		topo.validateNodes,
		topo.validateLinks,
	}
	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}
	return nil
}

// enrich populates missing fields in the original Topology struct.
func (topo *Topology) enrich() error {
	for name, node := range topo.Nodes {
		node.Name = name
		nodeLinks := make([]string, 0)
		for i, link := range topo.Links {
			// auto-assign link names
			link.Name = fmt.Sprintf("%s-%d", linkNamePrefix, i+1)
			for _, ep := range link.Endpoints {
				if !strings.Contains(ep, name) {
					continue
				}
				nodeLinks = append(nodeLinks, link.Name)
			}
		}
		node.Links = nodeLinks
	}
	return nil
}

// FromYAML validates and converts YAML data into a Topology struct.
func FromYAML(data []byte) (*Topology, error) {
	var topo Topology
	err := yaml.Unmarshal(data, &topo)
	if err != nil {
		return nil, errors.Join(ErrCorruptYAML, err)
	}
	if err := topo.validate(); err != nil {
		return nil, err
	}
	if err := topo.enrich(); err != nil {
		return nil, err
	}
	return &topo, nil
}
