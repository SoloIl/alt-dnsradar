package main

import "net"

func isPublicIPv4(ip net.IP) bool {
	if ip == nil {
		return false
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	switch {
	case ip4[0] == 10:
		return false
	case ip4[0] == 127:
		return false
	case ip4[0] == 0:
		return false
	case ip4[0] == 169 && ip4[1] == 254:
		return false
	case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
		return false
	case ip4[0] == 192 && ip4[1] == 168:
		return false
	case ip4[0] >= 224:
		return false
	case ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127:
		return false
	}

	return true
}
