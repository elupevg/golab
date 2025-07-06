package topology

import (
	"github.com/elupevg/golab/internal/vendors"
)

type Topology struct {
	Name          string           `yaml:"name"`
	Nodes         map[string]*Node `yaml:"nodes"`
	Links         []*Link          `yaml:"links"`
	ManageConfigs bool             `yaml:"manage_configs"`
}

type Node struct {
	Name          string
	Image         string   `yaml:"image"`
	Binds         []string `yaml:"binds"`
	Vendor        vendors.Vendor
	Interfaces    []*Interface
	IPv4Loopbacks []string `yaml:"ipv4_loopbacks"`
	IPv6Loopbacks []string `yaml:"ipv6_loopbacks"`
	Protocols     map[string]bool
	Sysctls       map[string]string
}

type Interface struct {
	Name       string
	Link       string
	IPv4Addr   string
	IPv6Addr   string
	DriverOpts map[string]string
}

type Link struct {
	Name        string   `yaml:"name"`
	Endpoints   []string `yaml:"endpoints"`
	IPv4Subnet  string   `yaml:"ipv4_subnet"`
	IPv6Subnet  string   `yaml:"ipv6_subnet"`
	IPv4Gateway string
	IPv6Gateway string
}
