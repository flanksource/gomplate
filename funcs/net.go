package funcs

import (
	"context"
	gonet "net"
)

// NetFuncs -
type NetFuncs struct {
	ctx context.Context
}

// CreateNetFuncs -
func CreateNetFuncs(ctx context.Context) map[string]interface{} {
	ns := &NetFuncs{ctx}
	return map[string]interface{}{
		"net": func() interface{} { return ns },
	}
}

// ContainsCIDR - reports whether ip is within the given CIDR block.
// Returns false on invalid input.
func (NetFuncs) ContainsCIDR(cidr, ip string) bool {
	_, network, err := gonet.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	addr := gonet.ParseIP(ip)
	return addr != nil && network.Contains(addr)
}

// IsValidIP - returns true if the string is a valid IPv4 or IPv6 address.
func (NetFuncs) IsValidIP(ip string) bool {
	return gonet.ParseIP(ip) != nil
}
