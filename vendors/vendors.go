// Package vendors provides vendor-specific customization tools.
package vendors

import "strings"

// Vendor represents a virtual network node vendor.
type Vendor int

const (
	UNKNOWN Vendor = iota
	FRR
	ARISTA
)

// DetectByImage attempts to detect a node vendor based on the container image name.
func DetectByImage(image string) Vendor {
	if strings.Contains(image, "frr") {
		return FRR
	}
	if strings.Contains(image, "ceos") {
		return ARISTA
	}
	return UNKNOWN
}
