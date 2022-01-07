package utils

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

func ByteCountToHuman(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPEZY"[exp])
}

func EstimateCuckooFilter(elementCount int64) uint {
	return uint(math.Ceil(-1 * float64(elementCount) * math.Log(0.03) / math.Pow(math.Log(2), 2)))
}

func UnsafeGetBytes(s string) []byte {
	return (*[0x7fff0000]byte)(unsafe.Pointer(
		(*reflect.StringHeader)(unsafe.Pointer(&s)).Data),
	)[:len(s):len(s)]
}

func ByteSliceToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}
