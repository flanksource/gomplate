package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetContainsCIDR(t *testing.T) {
	var f NetFuncs

	tests := []struct {
		cidr     string
		ip       string
		expected bool
	}{
		{"10.0.0.0/8", "10.1.2.3", true},
		{"10.0.0.0/8", "10.0.0.0", true},
		{"10.0.0.0/8", "10.255.255.255", true},
		{"10.0.0.0/8", "192.168.1.1", false},
		{"192.168.0.0/24", "192.168.0.50", true},
		{"192.168.0.0/24", "192.168.1.1", false},
		{"2001:db8::/32", "2001:db8::1", true},
		{"2001:db8::/32", "2001:db9::1", false},
		{"bad-cidr", "10.0.0.1", false},
		{"10.0.0.0/8", "not-an-ip", false},
		{"", "10.0.0.1", false},
		{"10.0.0.0/8", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.cidr+"/"+tc.ip, func(t *testing.T) {
			assert.Equal(t, tc.expected, f.ContainsCIDR(tc.cidr, tc.ip))
		})
	}
}

func TestNetIsValidIP(t *testing.T) {
	var f NetFuncs

	tests := []struct {
		ip       string
		expected bool
	}{
		{"10.0.0.1", true},
		{"192.168.1.255", true},
		{"0.0.0.0", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"not-an-ip", false},
		{"256.0.0.1", false},
		{"", false},
		{"10.0.0", false},
	}

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			assert.Equal(t, tc.expected, f.IsValidIP(tc.ip))
		})
	}
}
