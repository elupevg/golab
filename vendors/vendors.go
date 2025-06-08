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

// bindsByVendor stores collections of mandatory binds for Docker by vendor.
var bindsByVendor = map[Vendor][]string{
	FRR: {"/lib/modules:/lib/modules"},
}

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

// ExtraBinds returns a collection of mandatory binds for the specific vendor.
func ExtraBinds(v Vendor) []string {
	binds, ok := bindsByVendor[v]
	if !ok {
		return []string{}
	}
	return binds
}
