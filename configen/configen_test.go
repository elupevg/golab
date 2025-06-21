package configen_test

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/elupevg/golab/configen"
	"github.com/elupevg/golab/topology"
	"github.com/google/go-cmp/cmp"
)

const testYAML = `
name: "triangle"
generate_configs: true
ip_start_from:
  links:
    - 100.64.1.0/24
    - 2001:db8:64:1::/64
nodes:
  frr01:
    image: "quay.io/frrouting/frr:master"
  frr02:
    image: "quay.io/frrouting/frr:master"
  frr03:
    image: "quay.io/frrouting/frr:master"
links:
  - endpoints: ["frr01:eth0", "frr02:eth0"]
  - endpoints: ["frr01:eth1", "frr03:eth0"]
  - endpoints: ["frr02:eth1", "frr03:eth1"]
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
	err = configen.GenerateAndDump(topo, tempDir)
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
