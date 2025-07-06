package topology

import "github.com/goccy/go-yaml"

func parseYAML(data []byte) (*Topology, error) {
	var topo Topology
	err := yaml.Unmarshal(data, &topo)
	if err != nil {
		return nil, err
	}
	return &topo, nil
}

func FromYAML(data []byte) (*Topology, error) {
	topo, err := parseYAML(data)
	if err != nil {
		return nil, err
	}
	if err := topo.validate(); err != nil {
		return nil, err
	}
	if err := topo.populate(); err != nil {
		return nil, err
	}
	return topo, nil
}
