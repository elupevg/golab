// Package topology provides means to model a virtual network topology.
package topology

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/goccy/go-yaml"
)

const (
	linkNamePrefix       = "golab-link"
	defaultIPv4Range     = "100.64.0.0/24"
	defaultIPv4PrefixLen = 29
)

var (
	ErrCorruptYAML      = errors.New("cannot parse YAML file")
	ErrUnknownNode      = errors.New("unknown node in a link endpoint")
	ErrZeroNodes        = errors.New("topology has no nodes defined")
	ErrTooFewEndpoints  = errors.New("link has less than two endpoints")
	ErrInvalidEndpoint  = errors.New("invalid endpoint format")
	ErrInvalidCIDR      = errors.New("cannot parse IP")
	ErrInvalidInterface = errors.New("invalid interface name")
)

// Node represents a node in a virtual network topology.
type Node struct {
	Name       string
	Image      string   `yaml:"image"`
	Binds      []string `yaml:"binds"`
	Interfaces []*Interface
}

type Interface struct {
	Name string
	Link string
	IPv4 net.IP
}

// Link represents a link in a virtual network topology.
type Link struct {
	Endpoints     []string `yaml:"endpoints"`
	Name          string   `yaml:"name"`
	RawIPv4Subnet string   `yaml:"ipv4_subnet"`
	IPv4Subnet    *net.IPNet
	IPv4Gateway   net.IP
}

// Topology represents a virtual network comprised of nodes and links.
type Topology struct {
	Name      string           `yaml:"name"`
	Nodes     map[string]*Node `yaml:"nodes"`
	Links     []*Link          `yaml:"links"`
	IPv4Range *net.IPNet
}

// populate validates user-provided data and populates missing fields.
func (topo *Topology) populate() error {
	// check if at least one node is defined
	if len(topo.Nodes) == 0 {
		return ErrZeroNodes
	}
	// Prepare default IP pools
	_, ipv4Range, _ := net.ParseCIDR(defaultIPv4Range)
	topo.IPv4Range = ipv4Range

	for name, node := range topo.Nodes {
		node.Name = name
	}

	for i, link := range topo.Links {
		// auto-assign link name
		link.Name = fmt.Sprintf("%s-%d", linkNamePrefix, i+1)

		// Validate or assign IP subnet
		if link.RawIPv4Subnet == "" {
			link.IPv4Subnet, _ = cidr.NextSubnet(topo.IPv4Range, defaultIPv4PrefixLen)
		} else {
			_, ipv4net, err := net.ParseCIDR(link.RawIPv4Subnet)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidCIDR, link.RawIPv4Subnet)
			}
			link.IPv4Subnet = ipv4net
		}

		// Gateway is the last IP usable address of the subnet
		_, bcast := cidr.AddressRange(link.IPv4Subnet)
		link.IPv4Gateway = cidr.Dec(bcast)

		// check that link has at least two endpoints
		if len(link.Endpoints) < 2 {
			return fmt.Errorf("%w: %s", ErrTooFewEndpoints, link.Name)
		}
		for j, ep := range link.Endpoints {
			parts := strings.Split(ep, ":")
			if len(parts) != 2 {
				return fmt.Errorf("%w: %q", ErrInvalidEndpoint, ep)
			}
			nodeName, iface := parts[0], parts[1]
			node, ok := topo.Nodes[nodeName]
			if !ok {
				return fmt.Errorf("%w: %q in %q", ErrUnknownNode, nodeName, ep)
			}
			if !strings.HasPrefix(iface, "eth") {
				return fmt.Errorf("%w: %q in %q", ErrInvalidInterface, iface, ep)
			}
			ipv4Addr, err := cidr.Host(link.IPv4Subnet, j+1)
			if err != nil {
				return err
			}
			if node.Interfaces == nil {
				node.Interfaces = make([]*Interface, 0)
			}
			node.Interfaces = append(node.Interfaces, &Interface{
				Name: iface,
				Link: link.Name,
				IPv4: ipv4Addr,
			})
		}
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
	if err := topo.populate(); err != nil {
		return nil, err
	}
	return &topo, nil
}
