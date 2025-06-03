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
  - endpoints: ["frr01:eth2", "frr03:eth1"]
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

func (s *stubVirtProvider) NodeCreate(_ context.Context, _ topology.Node) error {
	if s.nodeErr != nil {
		return s.nodeErr
	}
	s.nodeCount++
	return nil
}

func TestBuild(t *testing.T) {
	t.Parallel()
	wantLinks, wantNodes := 2, 3
	vp := new(stubVirtProvider)
	err := golab.Build(context.Background(), []byte(testYAML), vp)
	if err != nil {
		t.Fatal(err)
	}
	if vp.nodeCount != wantNodes {
		t.Errorf("created nodes: want %d, got %d", wantNodes, vp.nodeCount)
	}
	if vp.linkCount != wantLinks {
		t.Errorf("created links: want %d, got %d", wantLinks, vp.linkCount)
	}
}

func TestBuildLinkError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to create link")
	vp := &stubVirtProvider{linkErr: wantErr}
	err := golab.Build(context.Background(), []byte(testYAML), vp)
	if !errors.Is(wantErr, err) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
}

func TestBuildNodeError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("failed to create node")
	vp := &stubVirtProvider{nodeErr: wantErr}
	err := golab.Build(context.Background(), []byte(testYAML), vp)
	if !errors.Is(wantErr, err) {
		t.Fatalf("error: want %q, got %q", wantErr, err)
	}
}

func TestBuildCorruptYAMLError(t *testing.T) {
	t.Parallel()
	err := golab.Build(context.Background(), []byte(`name`), new(stubVirtProvider))
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	wantErrMsg := "[1:1] string was used where mapping is expected\n>  1 | name\n       ^\n"
	if wantErrMsg != errMsg {
		t.Fatalf("error: want %q, got %q", wantErrMsg, errMsg)
	}
}

func TestWreck(t *testing.T) {
	t.Parallel()
	err := golab.Wreck(context.Background(), []byte(testYAML), new(stubVirtProvider))
	if err != nil {
		t.Fatal(err)
	}
}
