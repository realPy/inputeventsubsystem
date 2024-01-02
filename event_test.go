package inputeventsubsystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var data []byte = []byte{
	0x53, 0xaf, 0x93, 0x65, 0x0, 0x0, 0x0, 0x0, 0x61, 0xcf, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4, 0x0, 0x4, 0x0, 0x6, 0x0, 0x7, 0x0, 0x53, 0xaf, 0x93, 0x65, 0x0, 0x0, 0x0, 0x0, 0x61, 0xcf, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x2e, 0x0, 0x1, 0x0, 0x0, 0x0, 0x53, 0xaf, 0x93, 0x65, 0x0, 0x0, 0x0, 0x0, 0x61, 0xcf, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
}

func BenchmarkUnpackDeviceInputEvents(b *testing.B) {
	for it := 0; it < b.N; it++ {
		e := UnpackDeviceInputEvents(data)
		for _, ev := range e {
			eventPool.Put(ev)
		}
	}
}

func BenchmarkUnsafeUnpackDeviceInputEvents(b *testing.B) {
	for it := 0; it < b.N; it++ {
		UnsafeUnpackDeviceInputEvents(data)
	}
}
func TestUnit(t *testing.T) {
	e := UnpackDeviceInputEvents(data)

	e2 := UnsafeUnpackDeviceInputEvents(data)

	for index, ev := range e {
		assert.Equal(t, (e2)[index].Value, ev.Value)
		assert.Equal(t, (e2)[index].Type, ev.Type)
		assert.Equal(t, (e2)[index].Time, ev.Time)
		assert.Equal(t, (e2)[index].String(), ev.String())

	}

}
