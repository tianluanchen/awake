package pkg

import (
	"net"
)

func ResolveListenAddr(addr string) ([]string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if host != "0.0.0.0" && host != "::" && host != "" {
		return []string{addr}, nil
	}

	resolved := []string{"127.0.0.1:" + port}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, v := range addrs {
			ipNet, ok := v.(*net.IPNet)
			if ok {
				ipv4 := ipNet.IP.To4()
				if ipv4 != nil {
					resolved = append(resolved, ipv4.String()+":"+port)
				}
			}
		}
	}
	return resolved, nil
}
