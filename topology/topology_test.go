package topology_test

import (
	"errors"
	"testing"

	"github.com/elupevg/golab/topology"
	"github.com/google/go-cmp/cmp"
)

const goodYAML = `
name: "triangle"
nodes:
  frr01:
    image: "quay.io/frrouting/frr:master"
  frr02:
    image: "quay.io/frrouting/frr:master"
  frr03:
    image: "quay.io/frrouting/frr:master"
links:
  - endpoints: ["frr01:eth1", "frr02:eth1"]
    ipv4_subnet: 100.100.0.0/29
  - endpoints: ["frr01:eth2", "frr03:eth1"]
  - endpoints: ["frr02:eth2", "frr03:eth2"]
`

const badYAML = `name`

func TestTopologyFromYAML(t *testing.T) {
	t.Parallel()
	want := topology.Topology{
		Name: "triangle",
		Nodes: map[string]topology.Node{
			"frr01": {Image: "quay.io/frrouting/frr:master"},
			"frr02": {Image: "quay.io/frrouting/frr:master"},
			"frr03": {Image: "quay.io/frrouting/frr:master"},
		},
		Links: []topology.Link{
			{
				Endpoints:  []string{"frr01:eth1", "frr02:eth1"},
				IPv4Subnet: "100.100.0.0/29",
			},
			{Endpoints: []string{"frr01:eth2", "frr03:eth1"}},
			{Endpoints: []string{"frr02:eth2", "frr03:eth2"}},
		},
	}
	got, err := topology.FromYAML([]byte(goodYAML))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func TestTopologyFromYAMLError(t *testing.T) {
	t.Parallel()
	_, err := topology.FromYAML([]byte(badYAML))
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	wantErrMsg := "[1:1] string was used where mapping is expected\n>  1 | name\n       ^\n"
	if wantErrMsg != errMsg {
		t.Fatalf("error: want %q, got %q", wantErrMsg, errMsg)
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
			name: "UnknownNode",
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
