//go:build unix

package pkg

import (
	"os"
	"syscall"
)

func processExists(p *os.Process) bool {
	err := p.Signal(syscall.Signal(0))
	return err == nil
}
