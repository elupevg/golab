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
			want:  vendors.ARISTA,
		},
		{
			name:  "Unknown",
			image: "crpd:20.2R1.10",
			want:  vendors.UNKNOWN,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := vendors.DetectByImage(tc.image)
			if tc.want != got {
				t.Errorf("want %d, got %d", tc.want, got)
			}
		})
	}
}

func TestExtraBinds(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		vendor vendors.Vendor
		want   []string
	}{
		{
			name:   "FRRouting",
			vendor: vendors.FRR,
			want:   []string{"/lib/modules:/lib/modules"},
		},
		{
			name:   "Unknown",
			vendor: vendors.UNKNOWN,
			want:   []string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := vendors.ExtraBinds(tc.vendor)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
