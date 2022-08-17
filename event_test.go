package inputeventsubsystem

import "testing"

var data []byte = []byte{
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0x99, 0x98,
	0x99, 0x98,
	0xEE, 0xFF, 0x78, 0x78,
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0x99, 0x98,
	0x99, 0x98,
	0xEE, 0xFF, 0x78, 0x78,
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0x99, 0x98,
	0x99, 0x98,
	0xEE, 0xFF, 0x78, 0x78,
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0xEE, 0xFF, 0x78, 0x78, 0xEE, 0xFF, 0x78, 0x78,
	0x99, 0x98,
	0x99, 0x98,
	0xEE, 0xFF, 0x78, 0x78,
}

func BenchmarkUnpackDeviceInputEvents(b *testing.B) {
	e := UnpackDeviceInputEvents(data)
	for _, ev := range e {
		eventPool.Put(ev)
	}
}