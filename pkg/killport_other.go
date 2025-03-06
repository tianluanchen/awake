//go:build !unix

package pkg

import (
	"os"
)

func processExists(_ *os.Process) bool {
	return true
}
