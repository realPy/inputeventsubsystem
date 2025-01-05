//go:build !amd64 && !arm64
// +build !amd64,!arm64

package inputeventsubsystem

import "encoding/binary"

func GetTimevalValue(data []byte) int32 {
	return int32(binary.LittleEndian.Uint32(data))
}
