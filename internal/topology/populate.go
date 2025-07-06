package topology

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/elupevg/golab/internal/vendors"
)

const mplsLabels = 100_000

func (t *Topology) populate() error {
	for name, node := range t.Nodes {
		if err := node.populate(name); err != nil {
			return err
		}
	}
	for i, link := range t.Links {
		if err := link.populate(i, t.Nodes); err != nil {
			return err
		}
	}
	return nil
}

// populateBinds validates/fixes user provided binds and adds vendor-specific ones.
func (n *Node) populateBinds() {
	binds := make([]string, 0)
	extraBinds := vendors.ExtraBinds(n.Vendor)
	for _, bind := range n.Binds {
		paths := strings.Split(bind, ":")
		if !filepath.IsAbs(paths[0]) {
			// convert relative source path to absolute (UNIX only)
			paths[0] = filepath.Join(os.Getenv("PWD"), paths[0])
		}
		bind = strings.Join(paths, ":")
		if slices.Contains(extraBinds, bind) {
			continue
		}
		binds = append(binds, bind)
	}
	n.Binds = append(binds, extraBinds...)
}

func (n *Node) populate(name string) error {
	n.Name = name
	n.Vendor = vendors.DetectByImage(n.Image)
	if len(n.IPv4Loopbacks) == 0 {
		n.IPv4Loopbacks = []string{calcLoopback(name, 4)}
	}
	if len(n.IPv6Loopbacks) == 0 {
		n.IPv6Loopbacks = []string{calcLoopback(name, 6)}
	}
	if n.Vendor == vendors.FRR && n.Protocols["ldp"] {
		n.Sysctls = map[string]string{
			"net.mpls.platform_labels": strconv.Itoa(mplsLabels),
			"net.mpls.conf.lo.input":   "1",
		}
	}
	n.populateBinds()
	return nil
}

func calcLoopback(name string, ipVersion int) string {
	index, _ := strconv.Atoi(strings.TrimLeft(name, "R"))
	var loopback string
	switch ipVersion {
	case 4:
		loopback = fmt.Sprintf("192.168.0.%d/32", index)
	case 6:
		loopback = fmt.Sprintf("2001:db8::%d/128", index)
	}
	return loopback
}

func (l *Link) populate(i int, nodes map[string]*Node) error {
	l.Name = fmt.Sprintf("golab-link-%d", i+1)
	if l.IPv4Subnet == "" {
		l.IPv4Subnet = calcSubnet(l.Endpoints, 4)
	}
	if l.IPv6Subnet == "" {
		l.IPv6Subnet = calcSubnet(l.Endpoints, 6)
	}
	for _, ep := range l.Endpoints {
		node := nodes[ep]
		var driverOpts map[string]string
		if node.Vendor == vendors.FRR && node.Protocols["ldp"] {
			driverOpts = map[string]string{
				"com.docker.network.endpoint.sysctls": "net.mpls.conf.IFNAME.input=1",
			}
		}
		node.Interfaces = append(node.Interfaces, &Interface{
			Name:       "eth" + strconv.Itoa(len(node.Interfaces)),
			Link:       l.Name,
			IPv4Addr:   calcHost(l.IPv4Subnet, getIndex(ep)),
			IPv6Addr:   calcHost(l.IPv6Subnet, getIndex(ep)),
			DriverOpts: driverOpts,
		})
	}
	ipv4Gateway, _, _ := strings.Cut(calcHost(l.IPv4Subnet, 254), "/")
	ipv6Gateway, _, _ := strings.Cut(calcHost(l.IPv6Subnet, 254), "/")
	l.IPv4Gateway = ipv4Gateway
	l.IPv6Gateway = ipv6Gateway
	return nil
}

// calcSubnet generates a unique IP subnet based on the endpoints.
func calcSubnet(endpoints []string, ipVersion int) string {
	var a, b int
	if len(endpoints) > 2 {
		b = getIndex(endpoints[len(endpoints)-1])
	} else {
		a, b = getIndex(endpoints[0]), getIndex(endpoints[1])
	}
	var subnet string
	switch ipVersion {
	case 4:
		subnet = fmt.Sprintf("10.%d.%d.0/24", a, b)
	case 6:
		subnet = fmt.Sprintf("2001:db8:%d:%d::/64", a, b)
	}
	return subnet
}

func calcHost(subnet string, index int) string {
	net, pl, _ := strings.Cut(subnet, "/")
	if strings.Contains(subnet, ".") {
		net, _ = strings.CutSuffix(net, "0")
	}
	return fmt.Sprintf("%s%d/%s", net, index, pl)
}

// getIndex extracts a node index from the node name.
func getIndex(nodeName string) int {
	index, _ := strconv.Atoi(strings.TrimLeft(nodeName, "R"))
	return index
}
