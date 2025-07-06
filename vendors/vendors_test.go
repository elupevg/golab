package vendors_test

import (
	"testing"

	"github.com/elupevg/golab/vendors"
	"github.com/google/go-cmp/cmp"
)

func TestDetectByImage(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name  string
		image string
		want  vendors.Vendor
	}{
		{
			name:  "FRRouting",
			image: "quay.io/frrouting/frr:master",
			want:  vendors.FRR,
		},
		{
			name:  "Arista",
			image: "ceos:4.32.0F",
			want:  vendors.UNKNOWN,
		},
		{
			name:  "Juniper",
			image: "crpd:20.2R1.10",
			want:  vendors.UNKNOWN,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := vendors.DetectByImage(tc.image)
			if tc.want != got {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		vendor vendors.Vendor
		want   vendors.Config
	}{
		{
			name:   "FRRouting",
			vendor: vendors.FRR,
			want: vendors.Config{
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
		},
		{
			name:   "Unknown",
			vendor: vendors.UNKNOWN,
			want:   vendors.Config{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := vendors.GetConfig(tc.vendor)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
