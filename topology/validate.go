package topology

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

var supportedProtocols = map[string]bool{
	"bgp":   true,
	"ospf":  true,
	"ospf6": true,
	"isis":  true,
	"ldp":   true,
}

func (t *Topology) validate() error {
	if t.Name == "" {
		return errors.New("topology does not have a name")
	}
	if !t.ConfigMode.isValid() {
		return fmt.Errorf("topology %q has invalid config mode %q", t.Name, t.ConfigMode)
	}
	if len(t.Nodes) == 0 {
		return fmt.Errorf("topology %q has no nodes", t.Name)
	}
	nodeNames := make([]string, 0, len(t.Nodes))
	for name, node := range t.Nodes {
		if node == nil {
		}
		if err := node.validate(name); err != nil {
			return err
		}
		nodeNames = append(nodeNames, name)
	}
	for _, link := range t.Links {
		if err := link.validate(nodeNames); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) validate(name string) error {
	if n == nil {
		return fmt.Errorf("node %q does not have an image specified", name)
	}
	if !isValidNodeName(name) {
		return fmt.Errorf("node name %q is not in the [R1-R253] range", name)
	}
	if n.Image == "" {
		return fmt.Errorf("node %q does not have an image specified", name)
	}
	for _, bind := range n.Binds {
		if err := validateBind(bind); err != nil {
			return err
		}
	}
	for _, loop := range n.IPv4Loopbacks {
		if !isValidCIDR(loop, 4) {
			return fmt.Errorf("%q is not a valid IPv4 address", loop)
		}
	}
	for _, loop := range n.IPv6Loopbacks {
		if !isValidCIDR(loop, 6) {
			return fmt.Errorf("%q is not a valid IPv6 address", loop)
		}
	}
	for proto := range n.Protocols {
		if !supportedProtocols[proto] {
			return fmt.Errorf("node %q has unsupported protocol %q", name, proto)
		}
	}
	if n.ASN != nil && *(n.ASN) == 0 {
		return fmt.Errorf("node %q has unvalid ASN %d", name, *(n.ASN))
	}
	return nil
}

// isValidNodeName checks if the provided node name is compliant with the schema.
// Legal node names lie in the range R1..R253. This naming convention is enforced
// because node number is used in automated IP allocation. Number 254 is reserved
// for a gateway in each network, which is a requirement in Docker.
func isValidNodeName(name string) bool {
	after, found := strings.CutPrefix(name, "R")
	if !found {
		return false
	}
	num, err := strconv.Atoi(after)
	if err != nil {
		return false
	}
	return num > 0 && num < 254
}

// isValidCIDR tells if the provided string is a valid IP address in CIDR notation.
func isValidCIDR(s string, ipVersion int) bool {
	ip, _, err := net.ParseCIDR(s)
	if err != nil {
		return false
	}
	switch ipVersion {
	case 4:
		return ip.To4() != nil
	case 6:
		return ip.To4() == nil
	}
	return false
}

func validateBind(bind string) error {
	parts := strings.Split(bind, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid bind mount %q has invalid format", bind)
	}
	if !filepath.IsAbs(parts[1]) {
		return fmt.Errorf("invalid bind mount %q has non-absolute destination path", bind)
	}
	return nil
}

// validate runs sanity checks on the Link fields.
func (l *Link) validate(nodes []string) error {
	if len(l.Endpoints) < 2 {
		return fmt.Errorf("link has fewer than two endpoints %v", l.Endpoints)
	}
	for _, ep := range l.Endpoints {
		if !slices.Contains(nodes, ep) {
			return fmt.Errorf("unknown node %q in endpoints %v", ep, l.Endpoints)
		}
	}
	if l.IPv4Subnet != "" && !isValidCIDR(l.IPv4Subnet, 4) {
		return fmt.Errorf("%q is not a valid IPv4 subnet", l.IPv4Subnet)
	}
	if l.IPv6Subnet != "" && !isValidCIDR(l.IPv6Subnet, 6) {
		return fmt.Errorf("%q is not a valid IPv6 subnet", l.IPv6Subnet)
	}
	return nil
}

func (cm ConfigMode) isValid() bool {
	switch cm {
	case None, Manual, Auto:
		return true
	default:
		return false
	}
}
