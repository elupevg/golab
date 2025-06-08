package vendors_test

import (
	"testing"

	"github.com/elupevg/golab/vendors"
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
