package topology

import (
	"net"

	"github.com/elupevg/golab/internal/vendors"
)

// Topology represents a virtual network comprised of nodes and links.
type Topology struct {
	Name          string           `yaml:"name"`
	Nodes         map[string]*Node `yaml:"nodes"`
	Links         []*Link          `yaml:"links"`
	IPStartFrom   *IPStartFrom     `yaml:"ip_start_from"`
	ManageConfigs bool             `yaml:"manage_configs"`
}

// IPStartFrom represents a collection of initial subnets for auto-allocation.
type IPStartFrom struct {
	Links     []string `yaml:"links"`
	Loopbacks []string `yaml:"loopbacks"`
}

// Node represents a node in a virtual network topology.
type Node struct {
	Name       string
	Image      string   `yaml:"image"`
	Binds      []string `yaml:"binds"`
	Vendor     vendors.Vendor
	Interfaces []*Interface
	Loopbacks  []string `yaml:"loopbacks"`
	Protocols  map[string]string
	Enable     []string `yaml:"enable"`
	Sysctls    map[string]string
}

// Interface respresents a network node attachment to a link.
type Interface struct {
	Name       string
	Link       string
	IPv4       string
	IPv6       string
	DriverOpts map[string]string
}

// Link represents a link in a virtual network topology.
type Link struct {
	Name       string   `yaml:"name"`
	Endpoints  []string `yaml:"endpoints"`
	RawSubnets []string `yaml:"ip_subnets"`
	Subnets    []*net.IPNet
	Gateways   []net.IP
}
