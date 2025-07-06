package topology

import "testing"

func TestCalcSubnet(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		endpoints []string
		ipVersion int
		want      string
	}{
		{
			name:      "IPv4R3toR4",
			endpoints: []string{"R3", "R4"},
			ipVersion: 4,
			want:      "10.3.4.0/24",
		},
		{
			name:      "IPv6R3toR4",
			endpoints: []string{"R3", "R4"},
			ipVersion: 6,
			want:      "2001:db8:3:4::/64",
		},
		{
			name:      "IPv4R4toR3",
			endpoints: []string{"R4", "R3"},
			ipVersion: 4,
			want:      "10.4.3.0/24",
		},
		{
			name:      "IPv6R4toR3",
			endpoints: []string{"R4", "R3"},
			ipVersion: 6,
			want:      "2001:db8:4:3::/64",
		},
		{
			name:      "IPv4Broadcast",
			endpoints: []string{"R1", "R2", "R3"},
			ipVersion: 4,
			want:      "10.0.3.0/24",
		},
		{
			name:      "IPv6Broadcast",
			endpoints: []string{"R1", "R2", "R3"},
			ipVersion: 6,
			want:      "2001:db8:0:3::/64",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := calcSubnet(tc.endpoints, tc.ipVersion)
			if tc.want != got {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestCalcHost(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		subnet string
		index  int
		want   string
	}{
		{
			name:   "IPv4Host4",
			subnet: "10.3.4.0/24",
			index:  4,
			want:   "10.3.4.4/24",
		},
		{
			name:   "IPv6Host4",
			subnet: "2001:db8:3:4::/64",
			index:  4,
			want:   "2001:db8:3:4::4/64",
		},
		{
			name:   "IPv4Host254",
			subnet: "10.3.4.0/24",
			index:  254,
			want:   "10.3.4.254/24",
		},
		{
			name:   "IPv6Host254",
			subnet: "2001:db8:3:4::/64",
			index:  254,
			want:   "2001:db8:3:4::254/64",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := calcHost(tc.subnet, tc.index)
			if tc.want != got {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
