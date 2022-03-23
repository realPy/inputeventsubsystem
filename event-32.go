//go:build !amd64 && !arm64
// +build !amd64,!arm64

package inputeventsubsystem

func timeval(value uint32) int32 {

	return int32(value)
}
