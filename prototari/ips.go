package prototari

import (
	"errors"
	"fmt"
	"net"
)

// GetPrivateIPAndBroadcastAddr returns this computer's private IP (typically
// a 192.168.0.0/16 IP) and the broadcast IP in the private network.
// It iterates through the system network interfaces and stops when the first
// private IP is found. Thus if a computer is connected to the home wifi router
// using Ethernet and WIFI, one of the two will be chosen.
func GetPrivateIPAndBroadcastAddr() (net.IP, net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// Skip interfaces that are down, and loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			// Skip this interface: we can't get its addresses
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil && ipNet.IP.IsPrivate() {
				broadcastAddr := calculateBroadcastAddr(ipNet)
				return ipNet.IP, broadcastAddr, nil
			}
		}
	}

	return nil, nil, errors.New("no private IPv4 address found")
}

func calculateBroadcastAddr(ipNet *net.IPNet) net.IP {
	var (
		ip   = ipNet.IP.To4()
		mask = ipNet.Mask
	)

	if ip == nil {
		// Not an IPv4
		return nil
	}

	broadcast := make(net.IP, len(ip))
	for i := 0; i < len(ip); i++ {
		// Bitwise OR with the inverse of the mask
		broadcast[i] = ip[i] | ^mask[i]
	}

	return broadcast
}
