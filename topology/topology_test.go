package topology_test

import (
	"errors"
	"testing"

	"github.com/elupevg/golab/topology"
	"github.com/google/go-cmp/cmp"
)

const testYAML = `
name: "triangle"
nodes:
  frr01:
    image: "quay.io/frrouting/frr:master"
    binds:
      - "frr01:/etc/frr"
  frr02:
    image: "quay.io/frrouting/frr:master"
  frr03:
    image: "quay.io/frrouting/frr:master"
links:
  - endpoints: ["frr01:eth1", "frr02:eth1"]
    name: "golab-link-1"
    ipv4_subnet: 100.100.0.0/29
    ipv4_gateway: 100.100.0.6
  - endpoints: ["frr01:eth2", "frr03:eth1"]
    name: "golab-link-2"
  - endpoints: ["frr02:eth2", "frr03:eth2"]
    name: "golab-link-3"
`

func TestTopologyFromYAML(t *testing.T) {
	t.Parallel()
	want := topology.Topology{
		Name: "triangle",
		Nodes: map[string]topology.Node{
			"frr01": {
				Name:  "frr01",
				Image: "quay.io/frrouting/frr:master",
				Binds: []string{"frr01:/etc/frr"},
				Links: []string{"golab-link-1", "golab-link-2"},
			},
			"frr02": {
				Name:  "frr02",
				Image: "quay.io/frrouting/frr:master",
				Links: []string{"golab-link-1", "golab-link-3"},
			},
			"frr03": {
				Name:  "frr03",
				Image: "quay.io/frrouting/frr:master",
				Links: []string{"golab-link-2", "golab-link-3"},
			},
		},
		Links: []topology.Link{
			{
				Endpoints:   []string{"frr01:eth1", "frr02:eth1"},
				Name:        "golab-link-1",
				IPv4Subnet:  "100.100.0.0/29",
				IPv4Gateway: "100.100.0.6",
			},
			{
				Endpoints: []string{"frr01:eth2", "frr03:eth1"},
				Name:      "golab-link-2",
			},
			{
				Endpoints: []string{"frr02:eth2", "frr03:eth2"},
				Name:      "golab-link-3",
			},
		},
	}
	got, err := topology.FromYAML([]byte(testYAML))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestTopologyFromYAMLError(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		data   []byte
		errMsg string
	}{
		{
			name:   "CorruptYAML",
			data:   []byte(`name`),
			errMsg: "[1:1] string was used where mapping is expected\n>  1 | name\n       ^\n",
		},
		{
			name:   "ZeroNodesYAML",
			data:   []byte(`name: "triangle"`),
			errMsg: "topology has no nodes defined",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := topology.FromYAML(tc.data)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if tc.errMsg != errMsg {
				t.Fatalf("error: want %q, got %q", tc.errMsg, errMsg)
			}
		})
	}
}

func TestTopologyValidate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		topo topology.Topology
		err  error
	}{
		{
			name: "Success",
			topo: topology.Topology{
				Name: "triangle",
				Nodes: map[string]topology.Node{
					"frr01": {Image: "quay.io/frrouting/frr:master"},
					"frr02": {Image: "quay.io/frrouting/frr:master"},
					"frr03": {Image: "quay.io/frrouting/frr:master"},
				},
				Links: []topology.Link{
					{Endpoints: []string{"frr01:eth1", "frr02:eth1"}},
					{Endpoints: []string{"frr01:eth2", "frr03:eth1"}},
					{Endpoints: []string{"frr02:eth2", "frr03:eth2"}},
				},
			},
		},
		{
			name: "ZeroNodes",
			topo: topology.Topology{Name: "triangle"},
			err:  topology.ErrZeroNodes,
		},
		{
			name: "LinkWithOneEndpoint",
			topo: topology.Topology{
				Nodes: map[string]topology.Node{"frr01": {}},
				Links: []topology.Link{{Endpoints: []string{"frr01:eth1"}}},
			},
			err: topology.ErrTooFewEndpoints,
		},
		{
			name: "InvalidEndpoint",
			topo: topology.Topology{
				Nodes: map[string]topology.Node{"frr01": {}, "frr02": {}},
				Links: []topology.Link{{Endpoints: []string{"frr01:eth1", "frr02-eth1"}}},
			},
			err: topology.ErrInvalidEndpoint,
		},
		{
			name: "UnknownNode",
			topo: topology.Topology{
				Nodes: map[string]topology.Node{"frr01": {}, "frr02": {}},
				Links: []topology.Link{{Endpoints: []string{"frr01:eth1", "frr03:eth1"}}},
			},
			err: topology.ErrUnknownNode,
		},
		{
			name: "InvalidIPv4Subnet",
			topo: topology.Topology{
				Nodes: map[string]topology.Node{"frr01": {}, "frr02": {}},
				Links: []topology.Link{
					{
						Endpoints:  []string{"frr01:eth1", "frr02:eth1"},
						IPv4Subnet: "256.256.256.0/24",
					},
				},
			},
			err: topology.ErrInvalidIPv4Subnet,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.topo.Validate()
			if !errors.Is(tc.err, err) {
				t.Errorf("errors: want %q, got %q", tc.err, err)
			}
		})
	}
}
