//go:build arm64 || amd64
// +build arm64 amd64

package inputeventsubsystem

import "encoding/binary"

func GetTimevalValue(data []byte) int64 {
	return int64(binary.LittleEndian.Uint64(data))
}
