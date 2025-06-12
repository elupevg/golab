// Package topology provides means to model a virtual network topology.
package topology

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/elupevg/golab/vendors"
	"github.com/goccy/go-yaml"
)

const (
	autoLinkNamePrefix  = "golab-link-"
	autoIPv4FirstSubnet = "100.64.0.0/29"
	autoIPv4PrefixLen   = 29
	autoIPv6FirstSubnet = "2001:db8::/64"
	autoIPv6PrefixLen   = 64
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
	ErrMissingImage     = errors.New("node is missing image specification")
	ErrInvalidBind      = errors.New("invalid bind format")
)

// Node represents a node in a virtual network topology.
type Node struct {
	Name       string
	Image      string   `yaml:"image"`
	Binds      []string `yaml:"binds"`
	Interfaces []*Interface
	Vendor     vendors.Vendor
}

// Interface respresents a network node attachment to a link.
type Interface struct {
	Name string
	Link string
	IPv4 net.IP
	IPv6 net.IP
}

// Link represents a link in a virtual network topology.
type Link struct {
	Endpoints     []string `yaml:"endpoints"`
	Name          string   `yaml:"name"`
	RawIPv4Subnet string   `yaml:"ipv4_subnet"`
	IPv4Subnet    *net.IPNet
	IPv4Gateway   net.IP
	RawIPv6Subnet string `yaml:"ipv6_subnet"`
	IPv6Subnet    *net.IPNet
	IPv6Gateway   net.IP
}

// Topology represents a virtual network comprised of nodes and links.
type Topology struct {
	Name     string           `yaml:"name"`
	Nodes    map[string]*Node `yaml:"nodes"`
	Links    []*Link          `yaml:"links"`
	AutoIPv4 *net.IPNet
	AutoIPv6 *net.IPNet
}

// populateBinds validates/fixes user provided binds and adds vendor-specific ones.
func (n *Node) populateBinds() error {
	binds := make([]string, 0)
	extraBinds := vendors.ExtraBinds(n.Vendor)
	for _, bind := range n.Binds {
		// validate user defined binds
		paths := strings.Split(bind, ":")
		if len(paths) != 2 {
			// check that that bind contains source and target
			return fmt.Errorf("%w: %q", ErrInvalidBind, bind)
		}
		source, target := paths[0], paths[1]
		if !path.IsAbs(target) {
			// check whether target path is absolute
			return fmt.Errorf("%w: %q", ErrInvalidBind, bind)
		}
		if !path.IsAbs(source) {
			// convert relative source path to absolute (UNIX only)
			source = path.Join(os.Getenv("PWD"), source)
		}
		bind = source + ":" + target
		if slices.Contains(extraBinds, bind) {
			// ignore if user duplicated a vendor-specific bind
			continue
		}
		binds = append(binds, bind)
	}
	n.Binds = append(binds, extraBinds...)
	return nil
}

// populateNodes runs sanity checks on nodes and populates empty fields.
func (topo *Topology) populateNodes() error {
	// topology must contain at least one node
	if len(topo.Nodes) == 0 {
		return ErrZeroNodes
	}
	// populate node fields
	for name, node := range topo.Nodes {
		if node == nil || node.Image == "" {
			return fmt.Errorf("%w: %s", ErrMissingImage, name)
		}
		node.Name = name
		node.Vendor = vendors.DetectByImage(node.Image)
		if err := node.populateBinds(); err != nil {
			return err
		}
	}
	return nil
}

// populateLinks runs sanity checks on links and populates empty fields.
func (topo *Topology) populateLinks() error {
	for i, link := range topo.Links {
		// populate empty link names
		if link.Name == "" {
			link.Name = autoLinkNamePrefix + strconv.Itoa(i+1)
		}
		if link.RawIPv4Subnet == "" {
			// allocate next available IPv4 subnet
			link.IPv4Subnet = topo.AutoIPv4
			topo.AutoIPv4, _ = cidr.NextSubnet(topo.AutoIPv4, autoIPv4PrefixLen)
		} else {
			// validate IP subnet if manually allocated by the user
			_, ipv4net, err := net.ParseCIDR(link.RawIPv4Subnet)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidCIDR, link.RawIPv4Subnet)
			}
			link.IPv4Subnet = ipv4net
		}
		if link.RawIPv6Subnet == "" {
			// allocate next available IPv6 subnet
			link.IPv6Subnet = topo.AutoIPv6
			topo.AutoIPv6, _ = cidr.NextSubnet(topo.AutoIPv6, autoIPv6PrefixLen)
		} else {
			// validate IP subnet if manually allocated by the user
			_, ipv6net, err := net.ParseCIDR(link.RawIPv6Subnet)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidCIDR, link.RawIPv6Subnet)
			}
			link.IPv6Subnet = ipv6net
		}
		// allocate last usable IP address of the subnet as a gateway
		_, bcast := cidr.AddressRange(link.IPv4Subnet)
		link.IPv4Gateway = cidr.Dec(bcast)
		_, bcast = cidr.AddressRange(link.IPv6Subnet)
		link.IPv6Gateway = cidr.Dec(bcast)
		// check that link has at least two endpoints
		if len(link.Endpoints) < 2 {
			return fmt.Errorf("%w: %s", ErrTooFewEndpoints, link.Name)
		}
		for j, ep := range link.Endpoints {
			// validate endpoint string format
			parts := strings.Split(ep, ":")
			if len(parts) != 2 {
				return fmt.Errorf("%w: %q", ErrInvalidEndpoint, ep)
			}
			// map an endpoint to a node
			nodeName, iface := parts[0], parts[1]
			node, ok := topo.Nodes[nodeName]
			if !ok {
				return fmt.Errorf("%w: %q in %q", ErrUnknownNode, nodeName, ep)
			}
			// validate interface name
			if !strings.HasPrefix(iface, "eth") {
				return fmt.Errorf("%w: %q in %q", ErrInvalidInterface, iface, ep)
			}
			// allocate IP addresses
			ipv4Addr, err := cidr.Host(link.IPv4Subnet, j+1)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrSubnetExhausted, err)
			}
			ipv6Addr, err := cidr.Host(link.IPv6Subnet, j+1)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrSubnetExhausted, err)
			}
			node.Interfaces = append(node.Interfaces, &Interface{
				Name: iface,
				Link: link.Name,
				IPv4: ipv4Addr,
				IPv6: ipv6Addr,
			})
		}
	}
	return nil
}

// populateIPRanges runs sanity checks and initializes the IP pools for auto-allocation.
func (topo *Topology) populateIPRanges() error {
	_, autoIPnet, _ := net.ParseCIDR(autoIPv4FirstSubnet)
	topo.AutoIPv4 = autoIPnet
	_, autoIPnet, _ = net.ParseCIDR(autoIPv6FirstSubnet)
	topo.AutoIPv6 = autoIPnet
	return nil
}

// validator represents a generic validation check for a specific part of the topology.
type validator func() error

// populate runs sanity checks on the topology and populates empty fields.
func (topo *Topology) populate() error {
	validators := []validator{
		topo.populateNodes,
		topo.populateIPRanges,
		topo.populateLinks,
	}
	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
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
