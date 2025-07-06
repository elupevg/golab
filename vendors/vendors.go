// Package vendors provides vendor-specific configuration for network nodes.
package vendors

import "strings"

// Vendor represents a virtual network node vendor.
type Vendor string

const (
	UNKNOWN Vendor = ""
	FRR     Vendor = "frr"
)

// Config represents vendor-specific configuration for a node.
type Config struct {
	ImageSubstr string
	ConfigPath  string
	ConfigFiles []string
	ExtraBinds  []string
}

var configByVendor = map[Vendor]Config{
	FRR: {
		ImageSubstr: "frr",
		ConfigPath:  "/etc/frr",
		ConfigFiles: []string{
			"/etc/frr/daemons",
			"/etc/frr/vtysh.conf",
			"/etc/frr/frr.conf",
		},
		ExtraBinds: []string{
			"/lib/modules:/lib/modules",
		},
	},
}

// DetectByImage attempts to detect a node vendor based on the container image name.
func DetectByImage(image string) Vendor {
	for vendor, config := range configByVendor {
		if strings.Contains(image, config.ImageSubstr) {
			return vendor
		}
	}
	return UNKNOWN
}

// GetConfig provides vendor-specific configuration.
func GetConfig(v Vendor) Config {
	return configByVendor[v]
}
