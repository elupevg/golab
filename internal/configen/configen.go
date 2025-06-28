// Package configen provides means to generate network device configuration.
package configen

import (
	"embed"
	"errors"
	"html/template"
	"os"
	"path/filepath"

	"github.com/elupevg/golab/internal/logger"
	"github.com/elupevg/golab/internal/topology"
	"github.com/elupevg/golab/internal/vendors"
)

//go:embed templates
var configTemplates embed.FS

// ConfigenProvider stores cached logger.
type ConfigenProvider struct {
	log *logger.Logger
}

// New returns an instance of a ConfigenProvider.
func New(log *logger.Logger) *ConfigenProvider {
	return &ConfigenProvider{log}
}

// GenerateAndDump generates configs for all nodes in the topology and dumps them into provided directory.
func (cp *ConfigenProvider) GenerateAndDump(topo *topology.Topology, rootDir string) error {
	for _, node := range topo.Nodes {
		// create a directory for the node
		nodeDir := filepath.Join(rootDir, node.Name)
		err := os.Mkdir(nodeDir, 0o750)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				cp.log.Skipped("already created configuration for node " + node.Name)
				continue
			}
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
		cp.log.Success("generated configuration for node " + node.Name)
	}
	return nil
}

// Cleanup removes auto-generated configs for all nodes in the topology.
func (cp *ConfigenProvider) Cleanup(topo *topology.Topology, rootDir string) error {
	for _, node := range topo.Nodes {
		nodeDir := filepath.Join(rootDir, node.Name)
		_, err := os.Stat(nodeDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cp.log.Skipped("already removed configuration for node " + node.Name)
				continue
			}
			return err
		}
		if err := os.RemoveAll(nodeDir); err != nil {
			return err
		}
		cp.log.Success("removed configuration for node " + node.Name)
	}
	return nil
}
