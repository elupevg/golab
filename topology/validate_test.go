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

func TestIsValidNodeName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		want bool
	}{
		{
			name: "R1",
			want: true,
		},
		{
			name: "R253",
			want: true,
		},
		{
			name: "R254",
			want: false,
		},
		{
			name: "R0",
			want: false,
		},
		{
			name: "router2",
			want: false,
		},
		{
			name: "RXX2",
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidNodeName(tc.name)
			if tc.want != got {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLinkValidateErrors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		link   *Link
		errMsg string
	}{
		{
			name:   "Empty",
			link:   &Link{},
			errMsg: "link has fewer than two endpoints []",
		},
		{
			name:   "OneEndpoint",
			link:   &Link{Endpoints: []string{"R1"}},
			errMsg: "link has fewer than two endpoints [R1]",
		},
		{
			name:   "UnknownNode",
			link:   &Link{Endpoints: []string{"R1", "R9"}},
			errMsg: `unknown node "R9" in endpoints [R1 R9]`,
		},
		{
			name: "BadIPv4Subnet",
			link: &Link{
				Endpoints:  []string{"R1", "R2"},
				IPv4Subnet: "256.1.2.0/24",
				IPv6Subnet: "2001:db8:1:2::/64",
			},
			errMsg: `"256.1.2.0/24" is not a valid IPv4 subnet`,
		},
		{
			name: "BadIPv6Subnet",
			link: &Link{
				Endpoints:  []string{"R1", "R2"},
				IPv4Subnet: "10.1.2.0/24",
				IPv6Subnet: "2001:db8:1:2::/129",
			},
			errMsg: `"2001:db8:1:2::/129" is not a valid IPv6 subnet`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.link.validate([]string{"R1", "R2", "R3"})
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if tc.errMsg != errMsg {
				t.Errorf("want %q, got %q", tc.errMsg, errMsg)
			}
		})
	}
}
