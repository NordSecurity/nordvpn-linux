package cli

import (
	"fmt"
	"math/bits"
)

func uint64ToHumanBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	base := uint(bits.Len64(bytes) / 10)
	val := float64(bytes) / float64(uint64(1<<(base*10)))

	return fmt.Sprintf("%.2f %ciB", val, " KMGTPE"[base])
}
