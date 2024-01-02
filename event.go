package inputeventsubsystem

import (
	"encoding/binary"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

const deviceinputeventsize int = int(unsafe.Sizeof(Event{}))
const sizetimeval = int(unsafe.Sizeof(syscall.Timeval{}))

var eventPool = sync.Pool{
	New: func() interface{} { return new(Event) },
}

type Event struct {
	Time  syscall.Timeval // time in seconds since epoch at which event occurred
	Type  uint16          // event code related to the event type
	Code  uint16
	Value int32
}

func (ev Event) String() string {

	switch ev.Type {
	case EV_KEY:
		if strkeycode, ok := KeyCodesString[ev.Code]; ok {
			return fmt.Sprintf("{ time %d.%d, type %d (%s), code %d (%s), value %02d }",
				ev.Time.Sec, ev.Time.Usec, ev.Type, evtypeString[int(ev.Type)], ev.Code, strkeycode, ev.Value)
		} else {
			return fmt.Sprintf("{ time %d.%d, type %d (%s), code %d (%s), value %02d }",
				ev.Time.Sec, ev.Time.Usec, ev.Type, evtypeString[int(ev.Type)], ev.Code, BtnCodesString[ev.Code], ev.Value)
		}

	case EV_SYN:
		switch ev.Code {
		case SYN_REPORT:
			return fmt.Sprintf("{ time %d.%d ----------SYN REPORT-----------}", ev.Time.Sec, ev.Time.Usec)
		case SYN_DROPPED:
			return fmt.Sprintf("{ time %d.%d ----------++++++++++SYN DROPPED+++++++++-----------}", ev.Time.Sec, ev.Time.Usec)

		default:
			return fmt.Sprintf("{ time %d.%d, code %s, type %s, value %02d }",
				ev.Time.Sec, ev.Time.Usec, SynCodesString[ev.Code], evtypeString[int(ev.Type)], ev.Value)
		}

	case EV_REL:
		return fmt.Sprintf("{ time %d.%d, type %d (%s), code %d (%s), value %02d }",
			ev.Time.Sec, ev.Time.Usec, ev.Type, evtypeString[int(ev.Type)], ev.Code, RelCodesString[ev.Code], ev.Value)

	case EV_ABS:
		return fmt.Sprintf("{ time %d.%d, type %d (%s), code %d (%s), value %02d }",
			ev.Time.Sec, ev.Time.Usec, ev.Type, evtypeString[int(ev.Type)], ev.Code, AbsCodesString[ev.Code], ev.Value)

	default:
		return fmt.Sprintf("{ time %d.%d, code %02d, type %s, value %02d }",
			ev.Time.Sec, ev.Time.Usec, ev.Code, evtypeString[int(ev.Type)], ev.Value)
	}

}

func UnpackDeviceInputEvents(data []byte) []*Event {
	var events []*Event = make([]*Event, 0)
	var bytesconsum int = 0
	var i int = 0

	for {
		ev := eventPool.Get().(*Event)

		ev.Time.Sec = timeval(binary.LittleEndian.Uint32(data[i*deviceinputeventsize : i*deviceinputeventsize+4]))
		ev.Time.Usec = timeval(binary.LittleEndian.Uint32(data[i*deviceinputeventsize+sizetimeval-sizetimeval/2 : i*deviceinputeventsize+sizetimeval]))

		ev.Type = binary.LittleEndian.Uint16(data[i*deviceinputeventsize+sizetimeval : i*deviceinputeventsize+sizetimeval+2])
		ev.Code = binary.LittleEndian.Uint16(data[i*deviceinputeventsize+sizetimeval+2 : i*deviceinputeventsize+sizetimeval+4])
		ev.Value = int32(binary.LittleEndian.Uint32(data[i*deviceinputeventsize+sizetimeval+4 : i*deviceinputeventsize+sizetimeval+8]))

		events = append(events, ev)
		bytesconsum = bytesconsum + deviceinputeventsize
		if bytesconsum+deviceinputeventsize > len(data) {
			break
		}
		i++
	}
	return events
}

func UnsafeUnpackDeviceInputEvents(data []byte) []Event {
	ev := (*[]Event)(unsafe.Pointer(&data))
	return (*ev)[:len(data)/deviceinputeventsize]
}
