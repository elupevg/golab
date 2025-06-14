package topology_test

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/elupevg/golab/topology"
	"github.com/elupevg/golab/vendors"
	"github.com/google/go-cmp/cmp"
)

func TestFromYAML(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		data string
		want *topology.Topology
	}{
		{
			name: "ManualData",
			data: `
                        name: "triangle"
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                            binds: ["frr01:/etc/frr", "/lib/modules:/lib/modules"]
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                            binds: ["frr02:/etc/frr", "/lib/modules:/lib/modules"]
                          frr03:
                            image: "quay.io/frrouting/frr:master"
                            binds: ["frr03:/etc/frr", "/lib/modules:/lib/modules"]
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                            name: "ptp1"
                            ipv4_subnet: 100.64.1.0/29
                            ipv6_subnet: 2001:db8:1::/64
                          - endpoints: ["frr01:eth1", "frr03:eth0"]
                            name: "ptp2"
                            ipv4_subnet: 100.64.2.0/29
                            ipv6_subnet: 2001:db8:2::/64
                          - endpoints: ["frr02:eth1", "frr03:eth1"]
                            name: "ptp3"
                            ipv4_subnet: 100.64.3.0/29
                            ipv6_subnet: 2001:db8:3::/64
                        `,
			want: &topology.Topology{
				Name: "triangle",
				AutoIPv4: &net.IPNet{
					IP:   net.ParseIP("100.64.0.0"),
					Mask: net.CIDRMask(29, 32),
				},
				AutoIPv6: &net.IPNet{
					IP:   net.ParseIP("2001:db8::"),
					Mask: net.CIDRMask(64, 128),
				},
				Nodes: map[string]*topology.Node{
					"frr01": {
						Name:   "frr01",
						Vendor: vendors.FRR,
						Image:  "quay.io/frrouting/frr:master",
						Binds: []string{
							os.Getenv("PWD") + "/frr01:/etc/frr",
							"/lib/modules:/lib/modules",
						},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "ptp1",
								IPv4: net.ParseIP("100.64.1.1"),
								IPv6: net.ParseIP("2001:db8:1::1"),
							},
							{
								Name: "eth1",
								Link: "ptp2",
								IPv4: net.ParseIP("100.64.2.1"),
								IPv6: net.ParseIP("2001:db8:2::1"),
							},
						},
					},
					"frr02": {
						Name:   "frr02",
						Vendor: vendors.FRR,
						Image:  "quay.io/frrouting/frr:master",
						Binds: []string{
							os.Getenv("PWD") + "/frr02:/etc/frr",
							"/lib/modules:/lib/modules",
						},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "ptp1",
								IPv4: net.ParseIP("100.64.1.2"),
								IPv6: net.ParseIP("2001:db8:1::2"),
							},
							{
								Name: "eth1",
								Link: "ptp3",
								IPv4: net.ParseIP("100.64.3.1"),
								IPv6: net.ParseIP("2001:db8:3::1"),
							},
						},
					},
					"frr03": {
						Name:   "frr03",
						Vendor: vendors.FRR,
						Image:  "quay.io/frrouting/frr:master",
						Binds: []string{
							os.Getenv("PWD") + "/frr03:/etc/frr",
							"/lib/modules:/lib/modules",
						},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "ptp2",
								IPv4: net.ParseIP("100.64.2.2"),
								IPv6: net.ParseIP("2001:db8:2::2"),
							},
							{
								Name: "eth1",
								Link: "ptp3",
								IPv4: net.ParseIP("100.64.3.2"),
								IPv6: net.ParseIP("2001:db8:3::2"),
							},
						},
					},
				},
				Links: []*topology.Link{
					{
						Endpoints:     []string{"frr01:eth0", "frr02:eth0"},
						Name:          "ptp1",
						RawIPv4Subnet: "100.64.1.0/29",
						RawIPv6Subnet: "2001:db8:1::/64",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.1.0"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.1.6"),
						IPv6Subnet: &net.IPNet{
							IP:   net.ParseIP("2001:db8:1::"),
							Mask: net.CIDRMask(64, 128),
						},
						IPv6Gateway: net.ParseIP("2001:db8:1::ffff:ffff:ffff:fffe"),
					},
					{
						Endpoints:     []string{"frr01:eth1", "frr03:eth0"},
						Name:          "ptp2",
						RawIPv4Subnet: "100.64.2.0/29",
						RawIPv6Subnet: "2001:db8:2::/64",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.2.0"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.2.6"),
						IPv6Subnet: &net.IPNet{
							IP:   net.ParseIP("2001:db8:2::"),
							Mask: net.CIDRMask(64, 128),
						},
						IPv6Gateway: net.ParseIP("2001:db8:2::ffff:ffff:ffff:fffe"),
					},
					{
						Endpoints:     []string{"frr02:eth1", "frr03:eth1"},
						Name:          "ptp3",
						RawIPv4Subnet: "100.64.3.0/29",
						RawIPv6Subnet: "2001:db8:3::/64",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.3.0"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.3.6"),
						IPv6Subnet: &net.IPNet{
							IP:   net.ParseIP("2001:db8:3::"),
							Mask: net.CIDRMask(64, 128),
						},
						IPv6Gateway: net.ParseIP("2001:db8:3::ffff:ffff:ffff:fffe"),
					},
				},
			},
		},
		{
			name: "AutoPopulateData",
			data: `
                        name: "multihome"
                        nodes:
                          router:
                            image: "quay.io/frrouting/frr:master"
                          isp1:
                            image: "quay.io/frrouting/frr:master"
                          isp2:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["isp1:eth0", "router:eth0"]
                          - endpoints: ["isp2:eth0", "router:eth1"]
                        `,
			want: &topology.Topology{
				Name: "multihome",
				AutoIPv4: &net.IPNet{
					IP:   net.ParseIP("100.64.0.16"),
					Mask: net.CIDRMask(29, 32),
				},
				AutoIPv6: &net.IPNet{
					IP:   net.ParseIP("2001:db8:0:2::"),
					Mask: net.CIDRMask(64, 128),
				},
				Nodes: map[string]*topology.Node{
					"router": {
						Name:   "router",
						Vendor: vendors.FRR,
						Image:  "quay.io/frrouting/frr:master",
						Binds:  []string{"/lib/modules:/lib/modules"},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "golab-link-1",
								IPv4: net.ParseIP("100.64.0.2"),
								IPv6: net.ParseIP("2001:db8::2"),
							},
							{
								Name: "eth1",
								Link: "golab-link-2",
								IPv4: net.ParseIP("100.64.0.10"),
								IPv6: net.ParseIP("2001:db8:0:1::2"),
							},
						},
					},
					"isp1": {
						Name:   "isp1",
						Vendor: vendors.FRR,
						Image:  "quay.io/frrouting/frr:master",
						Binds:  []string{"/lib/modules:/lib/modules"},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "golab-link-1",
								IPv4: net.ParseIP("100.64.0.1"),
								IPv6: net.ParseIP("2001:db8::1"),
							},
						},
					},
					"isp2": {
						Name:   "isp2",
						Vendor: vendors.FRR,
						Image:  "quay.io/frrouting/frr:master",
						Binds:  []string{"/lib/modules:/lib/modules"},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "golab-link-2",
								IPv4: net.ParseIP("100.64.0.9"),
								IPv6: net.ParseIP("2001:db8:0:1::1"),
							},
						},
					},
				},
				Links: []*topology.Link{
					{
						Endpoints:     []string{"isp1:eth0", "router:eth0"},
						Name:          "golab-link-1",
						RawIPv4Subnet: "",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.0.0"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.0.6"),
						IPv6Subnet: &net.IPNet{
							IP:   net.ParseIP("2001:db8::"),
							Mask: net.CIDRMask(64, 128),
						},
						IPv6Gateway: net.ParseIP("2001:db8::ffff:ffff:ffff:fffe"),
					},
					{
						Endpoints:     []string{"isp2:eth0", "router:eth1"},
						Name:          "golab-link-2",
						RawIPv4Subnet: "",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.0.8"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.0.14"),
						IPv6Subnet: &net.IPNet{
							IP:   net.ParseIP("2001:db8:0:1::"),
							Mask: net.CIDRMask(64, 128),
						},
						IPv6Gateway: net.ParseIP("2001:db8:0:1:ffff:ffff:ffff:fffe"),
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := topology.FromYAML([]byte(tc.data))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestFromYAML_Errors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		data string
		err  error
	}{
		{
			name: "CorruptYAML",
			data: `
                        name
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                        `,
			err: topology.ErrCorruptYAML,
		},
		{
			name: "ZeroNodes",
			data: `name: "triangle"`,
			err:  topology.ErrZeroNodes,
		},
		{
			name: "LinkWithOneEndpoint",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0"]
                        `,
			err: topology.ErrTooFewEndpoints,
		},
		{
			name: "NodeWithoutFields",
			data: `
                        nodes:
                          frr01:
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                        `,
			err: topology.ErrMissingImage,
		},
		{
			name: "MissingImage",
			data: `
                        nodes:
                          frr01:
                            image: ""
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                        `,
			err: topology.ErrMissingImage,
		},
		{
			name: "InvalidEndpoint",
			data: `
                        nodes:
                          frr11:
                            image: "quay.io/frrouting/frr:master"
                          frr12:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr11:eth0", "frr12-eth1"]
                        `,
			err: topology.ErrInvalidEndpoint,
		},
		{
			name: "InvalidInterface",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:xe-0/0/0"]
                        `,
			err: topology.ErrInvalidInterface,
		},
		{
			name: "UnknownNode",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr03:eth0"]
                        `,
			err: topology.ErrUnknownNode,
		},
		{
			name: "InvalidIPv4Subnet",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                            ipv4_subnet: 256.0.0.0/29
                        `,
			err: topology.ErrInvalidCIDR,
		},
		{
			name: "InvalidIPv6Subnet",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                            ipv6_subnet: 2001:g::/64
                        `,
			err: topology.ErrInvalidCIDR,
		},
		{
			name: "IPv4SubnetExhausted",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                            ipv4_subnet: 100.64.0.0/31
                        `,
			err: topology.ErrSubnetExhausted,
		},
		{
			name: "IPv4SubnetExhausted",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                            ipv6_subnet: 2001:db8::/127
                        `,
			err: topology.ErrSubnetExhausted,
		},
		{
			name: "BindNonAbsTarget",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                            binds: ["/home/user/frr01:etc/frr"]
                        `,
			err: topology.ErrInvalidBind,
		},
		{
			name: "InvalidBindFormat",
			data: `
                        nodes:
                          frr01:
                            image: "quay.io/frrouting/frr:master"
                            binds: ["incorrect"]
                        `,
			err: topology.ErrInvalidBind,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := topology.FromYAML([]byte(tc.data))
			if !errors.Is(err, tc.err) {
				t.Errorf("error: want %q, got %q", tc.err, err)
			}
		})
	}
}
