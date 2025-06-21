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

const autoLinkNamePrefix = "golab-link-"

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
	ErrMissingSubnets   = errors.New("no subnets defined for a link")
)

// Topology represents a virtual network comprised of nodes and links.
type Topology struct {
	Name            string           `yaml:"name"`
	Nodes           map[string]*Node `yaml:"nodes"`
	Links           []*Link          `yaml:"links"`
	IPStartFrom     *IPStartFrom     `yaml:"ip_start_from"`
	GenerateConfigs bool             `yaml:"generate_configs"`
}

// IPStartFrom represents a collection of initial subnets for auto-allocation.
type IPStartFrom struct {
	RawLinks     []string `yaml:"links"`
	RawLoopbacks []string `yaml:"loopbacks"`
}

// Node represents a node in a virtual network topology.
type Node struct {
	Name       string
	Image      string   `yaml:"image"`
	Binds      []string `yaml:"binds"`
	Vendor     vendors.Vendor
	Interfaces []*Interface
}

// Interface respresents a network node attachment to a link.
type Interface struct {
	Name string
	Link string
	IPv4 string
	IPv6 string
}

// Link represents a link in a virtual network topology.
type Link struct {
	Name       string   `yaml:"name"`
	Endpoints  []string `yaml:"endpoints"`
	RawSubnets []string `yaml:"ip_subnets"`
	Subnets    []*net.IPNet
	Gateways   []net.IP
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

// autoSubnets calculates new set of subnets for the next link.
func (topo *Topology) autoSubnets() error {
	newSubnets := make([]string, 0, len(topo.IPStartFrom.RawLinks))
	for _, subnet := range topo.IPStartFrom.RawLinks {
		_, ipnet, err := net.ParseCIDR(subnet)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidCIDR, subnet)
		}
		prefixLen, _ := ipnet.Mask.Size()
		newIPNet, _ := cidr.NextSubnet(ipnet, prefixLen)
		newSubnets = append(newSubnets, newIPNet.String())
	}
	topo.IPStartFrom.RawLinks = newSubnets
	return nil
}

// allocateIPSubnets validates/allocates link IP subnets and addresses.
func (topo *Topology) allocateIPSubnets(link *Link) error {
	if link.RawSubnets == nil {
		if topo.IPStartFrom == nil || topo.IPStartFrom.RawLinks == nil {
			return fmt.Errorf("%w: %q", ErrMissingSubnets, link.Name)
		}
		link.RawSubnets = topo.IPStartFrom.RawLinks
		if err := topo.autoSubnets(); err != nil {
			return err
		}
	}
	for _, rawSubnet := range link.RawSubnets {
		// validate IP subnet if manually allocated by the user
		_, ipnet, err := net.ParseCIDR(rawSubnet)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidCIDR, rawSubnet)
		}
		link.Subnets = append(link.Subnets, ipnet)
		// allocate last usable IP address of the subnet as a gateway
		_, bcast := cidr.AddressRange(ipnet)
		link.Gateways = append(link.Gateways, cidr.Dec(bcast))
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
		if err := topo.allocateIPSubnets(link); err != nil {
			return err
		}
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
			var ipv4Addr, ipv6Addr string
			for _, subnet := range link.Subnets {
				// allocate IP addresses
				addr, err := cidr.Host(subnet, j+1)
				if err != nil {
					return fmt.Errorf("%w: %v", ErrSubnetExhausted, err)
				}
				pl, _ := subnet.Mask.Size()
				if ip := addr.To4(); ip != nil {
					ipv4Addr = addr.String() + "/" + strconv.Itoa(pl)
				} else {
					ipv6Addr = addr.String() + "/" + strconv.Itoa(pl)
				}
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

// populate runs sanity checks on the topology and populates empty fields.
func (topo *Topology) populate() error {
	validators := []func() error{
		topo.populateNodes,
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
