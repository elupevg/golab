package golab_test

import (
	"context"
	"errors"
	"testing"

	"github.com/elupevg/golab"
	"github.com/elupevg/golab/topology"
)

const testYAML = `
nodes:
  frr01:
    image: "quay.io/frrouting/frr:master"
  frr02:
    image: "quay.io/frrouting/frr:master"
  frr03:
    image: "quay.io/frrouting/frr:master"
links:
  - endpoints: ["frr01:eth1", "frr02:eth1"]
    ipv4_subnet: 100.64.1.0/29
    ipv4_gateway: 100.64.1.6
  - endpoints: ["frr01:eth2", "frr03:eth1"]
    ipv4_subnet: 100.64.2.0/29
    ipv4_gateway: 100.64.2.6
`

type stubVirtProvider struct {
	linkCount int
	nodeCount int
	linkErr   error
	nodeErr   error
}

func (s *stubVirtProvider) LinkCreate(_ context.Context, _ topology.Link) error {
	if s.linkErr != nil {
		return s.linkErr
	}
	s.linkCount++
	return nil
}

func (s *stubVirtProvider) LinkRemove(_ context.Context, _ topology.Link) error {
	if s.linkErr != nil {
		return s.linkErr
	}
	s.linkCount--
	return nil
}

func (s *stubVirtProvider) NodeCreate(_ context.Context, _ topology.Node) error {
	if s.nodeErr != nil {
		return s.nodeErr
	}
	s.nodeCount++
	return nil
}

func (s *stubVirtProvider) NodeRemove(_ context.Context, _ topology.Node) error {
	if s.nodeErr != nil {
		return s.nodeErr
	}
	s.nodeCount--
	return nil
}

func TestBuildWreck(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	vp := new(stubVirtProvider)
	// Build the topology
	wantLinks, wantNodes := 2, 3
	err := golab.Build(ctx, []byte(testYAML), vp)
	if err != nil {
		t.Fatal(err)
	}
	if vp.nodeCount != wantNodes {
		t.Fatalf("nodes: want %d, got %d", wantNodes, vp.nodeCount)
	}
	if vp.linkCount != wantLinks {
		t.Fatalf("links: want %d, got %d", wantLinks, vp.linkCount)
	}
	// Wreck the topology
	wantLinks, wantNodes = 0, 0
	err = golab.Wreck(ctx, []byte(testYAML), vp)
	if err != nil {
		t.Fatal(err)
	}
	if vp.nodeCount != wantNodes {
		t.Errorf("nodes: want %d, got %d", wantNodes, vp.nodeCount)
	}
	if vp.linkCount != wantLinks {
		t.Errorf("links: want %d, got %d", wantLinks, vp.linkCount)
	}
}

func TestBuildLinkError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to create link")
	vp := &stubVirtProvider{linkErr: wantErr}
	err := golab.Build(context.Background(), []byte(testYAML), vp)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
}

func TestBuildNodeError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to create node")
	vp := &stubVirtProvider{nodeErr: wantErr}
	err := golab.Build(context.Background(), []byte(testYAML), vp)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
}

func TestBuildCorruptYAMLError(t *testing.T) {
	t.Parallel()
	err := golab.Build(context.Background(), []byte(`name`), new(stubVirtProvider))
	if !errors.Is(err, topology.ErrCorruptYAML) {
		t.Fatalf("error: want %q, got %q", topology.ErrCorruptYAML, err)
	}
}

func TestWreckCorruptYAMLError(t *testing.T) {
	t.Parallel()
	err := golab.Wreck(context.Background(), []byte(`name`), new(stubVirtProvider))
	if !errors.Is(err, topology.ErrCorruptYAML) {
		t.Fatalf("error: want %q, got %q", topology.ErrCorruptYAML, err)
	}
}

func TestWreckLinkError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to remove link")
	vp := &stubVirtProvider{linkErr: wantErr}
	err := golab.Wreck(context.Background(), []byte(testYAML), vp)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
}

func TestWreckNodeError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to remove node")
	vp := &stubVirtProvider{nodeErr: wantErr}
	err := golab.Wreck(context.Background(), []byte(testYAML), vp)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
}
