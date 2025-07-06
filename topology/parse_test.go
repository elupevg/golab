package topology

import (
	"os"
	"testing"

	"github.com/elupevg/golab/vendors"
	"github.com/google/go-cmp/cmp"
)

func TestFromYAML(t *testing.T) {
	t.Parallel()
	testYAML := `
name: triangle
nodes:
  R1:
    image: "quay.io/frrouting/frr:master"
    binds: ["/lib/modules:/lib/modules"]
    protocols: {ldp: true}
  R2:
    image: "quay.io/frrouting/frr:master"
    binds: ["R2:/etc/frr"]
    protocols: {ospf: true, bgp: true}
    asn: 65000
  R3:
    image: "quay.io/frrouting/frr:master"
    ipv4_loopbacks: [172.16.0.3/32, 203.0.113.3/24]
    ipv6_loopbacks: [2001:db8:172:16::3/128, 2001:db8:203:113::3/64]
links:
  - endpoints: [R1, R2]
  - endpoints: [R1, R3]
  - endpoints: [R2, R3]
    ipv4_subnet: 100.64.0.0/24
    ipv6_subnet: 2001:db8:64::/64
`
	var testASN uint32 = 65000
	want := &Topology{
		Name: "triangle",
		Nodes: map[string]*Node{
			"R1": {
				Name:   "R1",
				Image:  "quay.io/frrouting/frr:master",
				Binds:  []string{"/lib/modules:/lib/modules"},
				Vendor: vendors.FRR,
				Interfaces: []*Interface{
					{
						Name:     "eth0",
						Link:     "golab-link-1",
						IPv4Addr: "10.1.2.1/24",
						IPv6Addr: "2001:db8:1:2::1/64",
						DriverOpts: map[string]string{
							"com.docker.network.endpoint.sysctls": "net.mpls.conf.IFNAME.input=1",
						},
					},
					{
						Name:     "eth1",
						Link:     "golab-link-2",
						IPv4Addr: "10.1.3.1/24",
						IPv6Addr: "2001:db8:1:3::1/64",
						DriverOpts: map[string]string{
							"com.docker.network.endpoint.sysctls": "net.mpls.conf.IFNAME.input=1",
						},
					},
				},
				IPv4Loopbacks: []string{"192.168.0.1/32"},
				IPv6Loopbacks: []string{"2001:db8::1/128"},
				Protocols:     map[string]bool{"ldp": true},
				Sysctls:       map[string]string{"net.mpls.conf.lo.input": "1", "net.mpls.platform_labels": "100000"},
			},
			"R2": {
				Name:  "R2",
				Image: "quay.io/frrouting/frr:master",
				Binds: []string{
					os.Getenv("PWD") + "/R2:/etc/frr",
					"/lib/modules:/lib/modules",
				},
				Vendor: vendors.FRR,
				Interfaces: []*Interface{
					{
						Name:     "eth0",
						Link:     "golab-link-1",
						IPv4Addr: "10.1.2.2/24", IPv6Addr: "2001:db8:1:2::2/64",
					},
					{
						Name:     "eth1",
						Link:     "golab-link-3",
						IPv4Addr: "100.64.0.2/24", IPv6Addr: "2001:db8:64::2/64",
					},
				},
				IPv4Loopbacks: []string{"192.168.0.2/32"},
				IPv6Loopbacks: []string{"2001:db8::2/128"},
				Protocols:     map[string]bool{"ospf": true, "bgp": true},
				ASN:           &testASN,
			},
			"R3": {
				Name:   "R3",
				Image:  "quay.io/frrouting/frr:master",
				Binds:  []string{"/lib/modules:/lib/modules"},
				Vendor: vendors.FRR,
				Interfaces: []*Interface{
					{
						Name:     "eth0",
						Link:     "golab-link-2",
						IPv4Addr: "10.1.3.3/24", IPv6Addr: "2001:db8:1:3::3/64",
					},
					{
						Name:     "eth1",
						Link:     "golab-link-3",
						IPv4Addr: "100.64.0.3/24", IPv6Addr: "2001:db8:64::3/64",
					},
				},
				IPv4Loopbacks: []string{
					"172.16.0.3/32",
					"203.0.113.3/24",
				},
				IPv6Loopbacks: []string{
					"2001:db8:172:16::3/128",
					"2001:db8:203:113::3/64",
				},
			},
		},
		Links: []*Link{
			{
				Name:        "golab-link-1",
				Endpoints:   []string{"R1", "R2"},
				IPv4Subnet:  "10.1.2.0/24",
				IPv6Subnet:  "2001:db8:1:2::/64",
				IPv4Gateway: "10.1.2.254",
				IPv6Gateway: "2001:db8:1:2::254",
			},
			{
				Name:        "golab-link-2",
				Endpoints:   []string{"R1", "R3"},
				IPv4Subnet:  "10.1.3.0/24",
				IPv6Subnet:  "2001:db8:1:3::/64",
				IPv4Gateway: "10.1.3.254",
				IPv6Gateway: "2001:db8:1:3::254",
			},
			{
				Name:        "golab-link-3",
				Endpoints:   []string{"R2", "R3"},
				IPv4Subnet:  "100.64.0.0/24",
				IPv6Subnet:  "2001:db8:64::/64",
				IPv4Gateway: "100.64.0.254",
				IPv6Gateway: "2001:db8:64::254",
			},
		},
	}
	got, err := FromYAML([]byte(testYAML))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}
