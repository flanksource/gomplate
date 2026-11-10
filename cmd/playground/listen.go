package main

import (
	"errors"
	"net"
	"syscall"
)

// listen retains the preferred port, or an OS-assigned port if it is occupied.
func listen(addr string, dynamic bool) (net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil || !dynamic || !errors.Is(err, syscall.EADDRINUSE) {
		return listener, err
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	return net.Listen("tcp", net.JoinHostPort(host, "0"))
}
