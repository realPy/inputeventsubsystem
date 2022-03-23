//go:build arm64 || amd64
// +build arm64 amd64

package inputeventsubsystem

func timeval(value uint32) int64 {

	return int64(value)
}
