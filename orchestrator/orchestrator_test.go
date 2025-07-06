package orchestrator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/elupevg/golab/orchestrator"
	"github.com/elupevg/golab/topology"
)

const testYAML = `
name: example
manage_configs: true
nodes:
  R1:
    image: "quay.io/frrouting/frr:master"
  R2:
    image: "quay.io/frrouting/frr:master"
  R3:
    image: "quay.io/frrouting/frr:master"
links:
  - endpoints: [R1, R2]
  - endpoints: [R1, R3]
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

type stubConfProvider struct {
	err error
}

func (s *stubConfProvider) GenerateAndDump(_ *topology.Topology, _ string) error {
	return s.err
}

func (s *stubConfProvider) Cleanup(_ *topology.Topology, _ string) error {
	return s.err
}

func TestBuildWreck(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	vp := new(stubVirtProvider)
	// build the topology
	wantLinks, wantNodes := 2, 3
	err := orchestrator.Build(ctx, []byte(testYAML), vp, new(stubConfProvider))
	if err != nil {
		t.Fatal(err)
	}
	if vp.nodeCount != wantNodes {
		t.Fatalf("nodes: want %d, got %d", wantNodes, vp.nodeCount)
	}
	if vp.linkCount != wantLinks {
		t.Fatalf("links: want %d, got %d", wantLinks, vp.linkCount)
	}
	// wreck the topology
	wantLinks, wantNodes = 0, 0
	err = orchestrator.Wreck(ctx, []byte(testYAML), vp, new(stubConfProvider))
	if err != nil {
		t.Fatal(err)
	}
	if vp.nodeCount != wantNodes {
		t.Fatalf("nodes: want %d, got %d", wantNodes, vp.nodeCount)
	}
	if vp.linkCount != wantLinks {
		t.Errorf("links: want %d, got %d", wantLinks, vp.linkCount)
	}
}

func TestBuildLinkError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to create link")
	vp := &stubVirtProvider{linkErr: wantErr}
	err := orchestrator.Build(context.Background(), []byte(testYAML), vp, new(stubConfProvider))
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestBuildNodeError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to create node")
	vp := &stubVirtProvider{nodeErr: wantErr}
	err := orchestrator.Build(context.Background(), []byte(testYAML), vp, new(stubConfProvider))
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestBuildCorruptYAMLError(t *testing.T) {
	t.Parallel()
	wantMsg := "[1:1] string was used where mapping is expected\n>  1 | name\n       ^\n"
	err := orchestrator.Build(context.Background(), []byte(`name`), new(stubVirtProvider), new(stubConfProvider))
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	if wantMsg != errMsg {
		t.Errorf("error: want %q, got %q", wantMsg, errMsg)
	}
}

func TestWreckCorruptYAMLError(t *testing.T) {
	t.Parallel()
	wantMsg := "[1:1] string was used where mapping is expected\n>  1 | name\n       ^\n"
	err := orchestrator.Wreck(context.Background(), []byte(`name`), new(stubVirtProvider), new(stubConfProvider))
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	if wantMsg != errMsg {
		t.Errorf("error: want %q, got %q", wantMsg, errMsg)
	}
}

func TestWreckLinkError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to remove link")
	vp := &stubVirtProvider{linkErr: wantErr}
	err := orchestrator.Wreck(context.Background(), []byte(testYAML), vp, new(stubConfProvider))
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestWreckNodeError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to remove node")
	vp := &stubVirtProvider{nodeErr: wantErr}
	err := orchestrator.Wreck(context.Background(), []byte(testYAML), vp, new(stubConfProvider))
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %q, got %q", wantErr, err)
	}
}

func TestBuildConfigError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	vp := new(stubVirtProvider)
	cp := new(stubConfProvider)
	cp.err = errors.New("failed to generate configs")
	wantLinks, wantNodes := 0, 0
	err := orchestrator.Build(ctx, []byte(testYAML), vp, cp)
	if !errors.Is(err, cp.err) {
		t.Fatalf("error: want %q, got %q", cp.err, err)
	}
	if vp.nodeCount != wantNodes {
		t.Fatalf("nodes: want %d, got %d", wantNodes, vp.nodeCount)
	}
	if vp.linkCount != wantLinks {
		t.Errorf("links: want %d, got %d", wantLinks, vp.linkCount)
	}
}

func TestWreckConfigError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	vp := new(stubVirtProvider)
	cp := new(stubConfProvider)
	err := orchestrator.Build(ctx, []byte(testYAML), vp, cp)
	if err != nil {
		t.Fatal(err)
	}
	cp.err = errors.New("failed to cleanup configs")
	wantLinks, wantNodes := 0, 0
	err = orchestrator.Wreck(ctx, []byte(testYAML), vp, cp)
	if !errors.Is(err, cp.err) {
		t.Fatalf("error: want %q, got %q", cp.err, err)
	}
	if vp.nodeCount != wantNodes {
		t.Fatalf("nodes: want %d, got %d", wantNodes, vp.nodeCount)
	}
	if vp.linkCount != wantLinks {
		t.Errorf("links: want %d, got %d", wantLinks, vp.linkCount)
	}
}
