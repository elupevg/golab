package configen_test

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/elupevg/golab/internal/configen"
	"github.com/elupevg/golab/internal/logger"
	"github.com/elupevg/golab/internal/topology"
	"github.com/google/go-cmp/cmp"
)

const testYAML = `
name: triangle
manage_configs: true
nodes:
  R1:
    image: "quay.io/frrouting/frr:master"
    protocols: {ospf: true, ospf6: true}
  R2:
    image: "quay.io/frrouting/frr:master"
  R3:
    image: "quay.io/frrouting/frr:master"
links:
  - endpoints: [R1, R2]
  - endpoints: [R1, R3]
  - endpoints: [R2, R3]
`

//go:embed testdata
var golden embed.FS

func files(fsys fs.FS) (paths []string) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		info, err := d.Info()
		if err != nil || info.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths
}

func TestGenerateAndDump(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	topo, err := topology.FromYAML([]byte(testYAML))
	if err != nil {
		t.Fatal(err)
	}
	cp := configen.New(logger.New(io.Discard, io.Discard))
	err = cp.GenerateAndDump(topo, tempDir)
	if err != nil {
		t.Fatal(err)
	}
	wantPaths := files(golden)
	for _, wantPath := range wantPaths {
		relPath, _ := filepath.Rel("testdata", wantPath)
		gotPath := filepath.Join(tempDir, relPath)
		// read generated config file
		got, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatalf("failed to open expected file %q", relPath)
		}
		// read golden config file
		want, err := golden.ReadFile(wantPath)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			t.Errorf("unexpected diff for %s: %s", relPath, diff)
		}
	}
}
