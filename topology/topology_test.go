package topology_test

import (
	"errors"
	"net"
	"testing"

	"github.com/elupevg/golab/topology"
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
                            binds: ["frr01:/etc/frr", "/lib/modules"]
                          frr02:
                            image: "quay.io/frrouting/frr:master"
                            binds: ["frr02:/etc/frr", "/lib/modules"]
                          frr03:
                            image: "quay.io/frrouting/frr:master"
                            binds: ["frr03:/etc/frr", "/lib/modules"]
                        links:
                          - endpoints: ["frr01:eth0", "frr02:eth0"]
                            name: "golab-link-1"
                            ipv4_subnet:  "100.64.1.0/29"
                            ipv4_gateway: "100.64.1.6"
                          - endpoints: ["frr01:eth1", "frr03:eth0"]
                            name: "golab-link-2"
                            ipv4_subnet:  "100.64.2.0/29"
                            ipv4_gateway: "100.64.2.6"
                          - endpoints: ["frr02:eth1", "frr03:eth1"]
                            name: "golab-link-3"
                            ipv4_subnet:  "100.64.3.0/29"
                            ipv4_gateway: "100.64.3.6"
                        `,
			want: &topology.Topology{
				Name: "triangle",
				IPv4Range: &net.IPNet{
					IP:   net.ParseIP("100.64.0.0"),
					Mask: net.CIDRMask(24, 32),
				},
				Nodes: map[string]*topology.Node{
					"frr01": {
						Name:  "frr01",
						Image: "quay.io/frrouting/frr:master",
						Binds: []string{"frr01:/etc/frr", "/lib/modules"},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "golab-link-1",
								IPv4: net.ParseIP("100.64.1.1"),
							},
							{
								Name: "eth1",
								Link: "golab-link-2",
								IPv4: net.ParseIP("100.64.2.1"),
							},
						},
					},
					"frr02": {
						Name:  "frr02",
						Image: "quay.io/frrouting/frr:master",
						Binds: []string{"frr02:/etc/frr", "/lib/modules"},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "golab-link-1",
								IPv4: net.ParseIP("100.64.1.2"),
							},
							{
								Name: "eth1",
								Link: "golab-link-3",
								IPv4: net.ParseIP("100.64.3.1"),
							},
						},
					},
					"frr03": {
						Name:  "frr03",
						Image: "quay.io/frrouting/frr:master",
						Binds: []string{"frr03:/etc/frr", "/lib/modules"},
						Interfaces: []*topology.Interface{
							{
								Name: "eth0",
								Link: "golab-link-2",
								IPv4: net.ParseIP("100.64.2.2"),
							},
							{
								Name: "eth1",
								Link: "golab-link-3",
								IPv4: net.ParseIP("100.64.3.2"),
							},
						},
					},
				},
				Links: []*topology.Link{
					{
						Endpoints:     []string{"frr01:eth0", "frr02:eth0"},
						Name:          "golab-link-1",
						RawIPv4Subnet: "100.64.1.0/29",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.1.0"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.1.6"),
					},
					{
						Endpoints:     []string{"frr01:eth1", "frr03:eth0"},
						Name:          "golab-link-2",
						RawIPv4Subnet: "100.64.2.0/29",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.2.0"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.2.6"),
					},
					{
						Endpoints:     []string{"frr02:eth1", "frr03:eth1"},
						Name:          "golab-link-3",
						RawIPv4Subnet: "100.64.3.0/29",
						IPv4Subnet: &net.IPNet{
							IP:   net.ParseIP("100.64.3.0"),
							Mask: net.CIDRMask(29, 32),
						},
						IPv4Gateway: net.ParseIP("100.64.3.6"),
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
                          frr02:
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
                            ipv4_subnet: "256.0.0.0/29"
                        `,
			err: topology.ErrInvalidCIDR,
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
