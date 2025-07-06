package topology

import "testing"

func TestIsValidCIDR(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		addr string
		ipv  int
		want bool
	}{
		{
			name: "IPv4Addr",
			addr: "10.3.4.3/24",
			ipv:  4,
			want: true,
		},
		{
			name: "IPv6Addr",
			addr: "2001:db8:3:4::3/64",
			ipv:  6,
			want: true,
		},
		{
			name: "IPv4Net",
			addr: "10.3.4.0/24",
			ipv:  4,
			want: true,
		},
		{
			name: "IPv6Net",
			addr: "2001:db8:3:4::/64",
			ipv:  6,
			want: true,
		},
		{
			name: "IPv6InsteadOfIPv4",
			addr: "2001:db8:3:4::3/64",
			ipv:  4,
			want: false,
		},
		{
			name: "IPv4InsteadOfIPv6",
			addr: "10.3.4.3/24",
			ipv:  6,
			want: false,
		},
		{
			name: "NonIPv4",
			addr: "something",
			ipv:  4,
			want: false,
		},
		{
			name: "NonIPv6",
			addr: "something",
			ipv:  6,
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidCIDR(tc.addr, tc.ipv)
			if tc.want != got {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}
