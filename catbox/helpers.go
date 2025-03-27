package catbox

import (
	"math/rand/v2"
	"time"
)

// Without extension
func RandomName(n int) string {
	b := make([]byte, n)
	for i := range n {
		v := rand.IntN(52)
		if v >= 26 {
			v -= 26
			b[i] = byte(97 + v)
		} else {
			b[i] = byte(65 + v)
		}
	}
	return string(b)
}

func IsUnsupportedExtension(ext string) bool {
	switch ext {
	case ".exe", ".scr", ".cpl", ".jar":
	case ".doc", ".docx", ".docm", ".docb":
	default:
		return false
	}
	return true
}

func IsValidStorageDuration(d time.Duration) bool {
	return d == 0 || d == time.Hour || d == time.Hour*12 || d == time.Hour*24 || d == time.Hour*72
}

func MaxUploadSize(d time.Duration) int64 {
	switch d {
	case 0:
		return 200 * 1024 * 1024
	case time.Hour, time.Hour * 12, time.Hour * 24, time.Hour * 72:
		return 1024 * 1024 * 1024
	default:
		return 0
	}
}
