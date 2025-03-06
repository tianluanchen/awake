package pkg

import (
	"errors"
	"os"

	gnet "github.com/shirou/gopsutil/v4/net"
)

// kind must be one of "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "inet", "inet4", "inet6"
//
// default signal is os.Kill
func KillPortProcess(port int, kind string) error {
	p := uint32(port)
	switch kind {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "inet", "inet4", "inet6":
	default:
		return errors.New("kind must be one of tcp, tcp4, tcp6, udp, udp4, udp6, inet, inet4, inet6")
	}
	connections, err := gnet.Connections(kind)
	if err != nil {
		return err
	}
	pidMap := make(map[int]struct{})
	for _, conn := range connections {
		if conn.Laddr.Port == p {
			pid := int(conn.Pid)
			if _, ok := pidMap[pid]; ok {
				continue
			}
			pidMap[pid] = struct{}{}
			process, err := os.FindProcess(pid)
			if err != nil {
				continue
			}
			if !processExists(process) {
				continue
			}
			err = process.Signal(os.Kill)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
