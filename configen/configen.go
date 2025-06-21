// Package configen provides means to generate network devices' configuration.
package configen

import (
	"embed"
	"html/template"
	"os"
	"path/filepath"

	"github.com/elupevg/golab/topology"
	"github.com/elupevg/golab/vendors"
)

//go:embed templates
var configTemplates embed.FS

func GenerateAndDump(topo *topology.Topology, rootDir string) error {
	for _, node := range topo.Nodes {
		// create a directory for the node
		nodeDir := filepath.Join(rootDir, node.Name)
		if err := os.Mkdir(nodeDir, 0o750); err != nil {
			return err
		}
		var remoteDir, fileName string
		for _, path := range vendors.ConfigFiles(node.Vendor) {
			remoteDir, fileName = filepath.Split(path)
			// prepare a template
			tmplData, err := configTemplates.ReadFile(filepath.Join("templates", node.Vendor.String(), fileName+".tmpl"))
			if err != nil {
				return err
			}
			tmpl, err := template.New(fileName).Parse(string(tmplData))
			if err != nil {
				return err
			}
			// render config and dump it into a file
			f, err := os.Create(filepath.Join(nodeDir, fileName))
			if err != nil {
				return err
			}
			defer f.Close()
			if err := tmpl.Execute(f, node); err != nil {
				return err
			}
		}
		node.Binds = append(node.Binds, nodeDir+":"+remoteDir)
	}
	return nil
}
