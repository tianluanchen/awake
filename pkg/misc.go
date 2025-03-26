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

func IsIP(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil
}

// https://github.com/golang/go/blob/9ce47e66e8eb7348005f1e45a77a729e7f146a1e/src/net/dnsclient.go#L89
func IsDomainName(s string) bool {
	// The root domain name is valid. See golang.org/issue/45715.
	if s == "." {
		return true
	}

	// See RFC 1035, RFC 3696.
	// Presentation format has dots before every label except the first, and the
	// terminal empty label is optional here because we assume fully-qualified
	// (absolute) input. We must therefore reserve space for the first and last
	// labels' length octets in wire format, where they are necessary and the
	// maximum total length is 255.
	// So our _effective_ maximum is 253, but 254 is not rejected if the last
	// character is a dot.
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
			nonNumeric = true
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return nonNumeric
}

func IsValidPort(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}
	return true
}
