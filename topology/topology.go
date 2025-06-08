// Package topology provides means to model a virtual network topology.
package topology

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/goccy/go-yaml"
)

const (
	autoLinkNamePrefix  = "golab-link-"
	autoIPv4FirstSubnet = "100.64.0.0/29"
	autoIPv4PrefixLen   = 29
)

var (
	ErrCorruptYAML      = errors.New("cannot parse YAML file")
	ErrUnknownNode      = errors.New("unknown node in a link endpoint")
	ErrZeroNodes        = errors.New("topology has no nodes defined")
	ErrTooFewEndpoints  = errors.New("link has less than two endpoints")
	ErrInvalidEndpoint  = errors.New("invalid endpoint format")
	ErrInvalidCIDR      = errors.New("cannot parse IP")
	ErrInvalidInterface = errors.New("invalid interface name")
	ErrSubnetExhausted  = errors.New("cannot allocate IP address")
)

// Node represents a node in a virtual network topology.
type Node struct {
	Name       string
	Image      string   `yaml:"image"`
	Binds      []string `yaml:"binds"`
	Interfaces []*Interface
}

// Interface respresents a network node attachment to a link.
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
	Name     string           `yaml:"name"`
	Nodes    map[string]*Node `yaml:"nodes"`
	Links    []*Link          `yaml:"links"`
	AutoIPv4 *net.IPNet
}

// populate validates data provided by the user and auto-populates missing fields.
func (topo *Topology) populate() error {
	// check if at least one node is defined
	if len(topo.Nodes) == 0 {
		return ErrZeroNodes
	}
	// initialize IP pools for auto-allocation
	_, autoIPv4net, _ := net.ParseCIDR(autoIPv4FirstSubnet)
	topo.AutoIPv4 = autoIPv4net
	// populate node names
	for name, node := range topo.Nodes {
		node.Name = name
	}
	for i, link := range topo.Links {
		// populate missing link names
		if link.Name == "" {
			link.Name = autoLinkNamePrefix + strconv.Itoa(i+1)
		}
		if link.RawIPv4Subnet == "" {
			// auto-assign next available IP subnet
			link.IPv4Subnet = topo.AutoIPv4
			topo.AutoIPv4, _ = cidr.NextSubnet(topo.AutoIPv4, autoIPv4PrefixLen)
		} else {
			// or validate manually assigned IP subnet
			_, ipv4net, err := net.ParseCIDR(link.RawIPv4Subnet)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidCIDR, link.RawIPv4Subnet)
			}
			link.IPv4Subnet = ipv4net
		}
		// gateway is the last usable IP address of the subnet
		_, bcast := cidr.AddressRange(link.IPv4Subnet)
		link.IPv4Gateway = cidr.Dec(bcast)
		// check that link has at least two endpoints
		if len(link.Endpoints) < 2 {
			return fmt.Errorf("%w: %s", ErrTooFewEndpoints, link.Name)
		}
		for j, ep := range link.Endpoints {
			// validate sanity of each endpoint and map them to nodes
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
				return fmt.Errorf("%w: %v", ErrSubnetExhausted, err)
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
