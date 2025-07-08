package topology

import (
	"github.com/elupevg/golab/vendors"
)

type IPMode string

const (
	Unknown IPMode = ""
	IPv4    IPMode = "ipv4"
	IPv6    IPMode = "ipv6"
	Dual    IPMode = "dual"
)

type ConfigMode string

const (
	None   ConfigMode = ""
	Manual ConfigMode = "manual"
	Auto   ConfigMode = "auto"
)

type Topology struct {
	Name       string           `yaml:"name"`
	Nodes      map[string]*Node `yaml:"nodes"`
	Links      []*Link          `yaml:"links"`
	ConfigMode ConfigMode       `yaml:"config_mode"`
	IPMode     IPMode           `yaml:"ip_mode"`
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
	ASN           *uint32
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
