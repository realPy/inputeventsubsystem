package inputeventsubsystem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type AbsInfo struct {
	Value      int32
	Minimum    int32
	Maximum    int32
	Fuzz       int32
	Flat       int32
	Resolution int32
}

func (a *AbsInfo) Unpack(data []byte) {

	//very important , we used basic read . Read with a struct use reflection and are very slow

	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &a.Value)
	binary.Read(buf, binary.LittleEndian, &a.Minimum)
	binary.Read(buf, binary.LittleEndian, &a.Maximum)
	binary.Read(buf, binary.LittleEndian, &a.Fuzz)
	binary.Read(buf, binary.LittleEndian, &a.Flat)
	binary.Read(buf, binary.LittleEndian, &a.Resolution)
}

type Device struct {
	Fn            string   // path to input device (devnode)
	File          *os.File // an open file handle to the input device
	fd            int
	DriverVersion uint32
	bus           uint16
	VendorID      uint16
	ProductID     uint16
	Version       uint16
	Name          string
	Phy           string
	Capabilities  map[int]map[int]string
	Absinfos      map[int]AbsInfo
	eventchan     chan []*Event
	errorchan     chan error
	stopped       int32
}

func (e *Device) String() string {
	return fmt.Sprintf("%s: bus 0x%x vendor 0x%x product 0x%x version 0x%x\n", e.Name, e.bus, e.VendorID, e.ProductID, e.VendorID)
}

// Open an evdev input device.
func Open(devnode string, buffersize int) (*Device, error) {

	var dev Device
	dev.Fn = devnode

	f, err := unix.Open(dev.Fn, syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0666)
	//f, err := syscall.Open(dev.Fn, syscall.O_NONBLOCK, 0666)

	if err != nil {
		return nil, err
	}

	dev.fd = f

	if dev.DriverVersion, err = IoctlInputVersion(dev.fd); err != nil {

		dev.DriverVersion = 0
		defer syscall.Close(dev.fd)
		return nil, ErrDriverVersion
	}

	if dev.bus, dev.VendorID, dev.ProductID, dev.Version, err = IoctlInputID(dev.fd); err != nil {
		defer syscall.Close(dev.fd)
		return nil, ErrDeviceInformation
	}

	dev.eventchan = make(chan []*Event, buffersize)
	dev.errorchan = make(chan error)

	dev.Name, _ = IoctlInputName(dev.fd)
	dev.Phy, _ = IoctlInputPhys(dev.fd)

	var evbits []byte

	if evbits, err = IoctlInputBit(dev.fd, 0, EV_MAX); err != nil {
		defer syscall.Close(dev.fd)
		return nil, ErrEvBits

	}

	dev.Capabilities = make(map[int]map[int]string)
	dev.Absinfos = make(map[int]AbsInfo)

	for evtype := 0; evtype < EV_MAX; evtype++ {
		if evbits[evtype/8]&(1<<uint(evtype%8)) != 0 {

			dev.Capabilities[evtype] = make(map[int]string)

			if evtype == EV_KEY {

				var codebits []byte

				if codebits, err = IoctlInputBit(dev.fd, evtype, KEY_MAX); err == nil {

					for evcode := 0; evcode < KEY_MAX; evcode++ {
						if codebits[evcode/8]&(1<<uint(evcode%8)) != 0 {
							dev.Capabilities[evtype][evcode] = fmt.Sprintf("0x%x", evcode)

						}
					}

				}

			}

			if evtype == EV_ABS {

				var absbits []byte
				if absbits, err = IoctlInputBit(dev.fd, evtype, ABS_MAX); err == nil {

					for abscode := 0; abscode < ABS_MAX; abscode++ {
						if absbits[abscode/8]&(1<<uint(abscode%8)) != 0 {

							dev.Capabilities[evtype][abscode] = fmt.Sprintf("0x%x", abscode)

							//hat not have absinfo
							if abscode < ABS_HAT0X || abscode > ABS_HAT3Y {

								var absinfobits []byte

								if absinfobits, err = IoctlInputAbs(dev.fd, abscode); err == nil {
									var a AbsInfo

									a.Unpack(absinfobits[:])
									dev.Absinfos[abscode] = a

								}

							}

						}

					}

				}

			}

		}
	}

	return &dev, nil
}

func (dev *Device) Error() <-chan error {

	return dev.errorchan
}

func (dev *Device) Read() chan []*Event {

	go func() {
		var events [deviceinputeventsize * 64]byte

		for {

			rFdSet := &unix.FdSet{}
			fd := int(dev.fd)
			rFdSet.Set(fd)

			t := unix.Timespec{Sec: 1 /*sec*/, Nsec: 0 /*usec*/}

			if _, err := unix.Pselect(fd+1, rFdSet, nil, nil, &t, nil); err == nil {

				if n, err := unix.Read(fd, events[:]); err == nil {
					p := UnpackDeviceInputEvents(events[0:n])
					dev.eventchan <- p

				} else {

					if err != syscall.EWOULDBLOCK {
						select {
						case dev.errorchan <- err:

						case <-time.After(time.Duration(100) * time.Millisecond):
						}
						return
					}

				}

			}

			if atomic.LoadInt32(&dev.stopped) == 1 {
				return
			}

		}

	}()
	return dev.eventchan
}

func (dev *Device) Grab(state bool) error {
	return IoctlInputGrab(dev.fd, state)
}

func (dev *Device) StopRead() {
	dev.Close()
}

func (dev *Device) ReadDone(events []*Event) {

	for _, ev := range events {
		eventPool.Put(ev)
	}
}

func (dev *Device) Close() error {
	atomic.StoreInt32(&dev.stopped, 1)
	return syscall.Close(dev.fd)
}

func (dev *Device) KeysState() ([]byte, error) {

	var keybits []byte
	var err error

	if keybits, err = IoctlInputKey(dev.fd); err == nil {
		return keybits, nil
	}

	return nil, ErrEvBits

}

func (dev *Device) AbsState(abscode int) (AbsInfo, error) {
	var a AbsInfo

	var absinfobits []byte
	var err error

	if absinfobits, err = IoctlInputAbs(dev.fd, abscode); err == nil {
		var a AbsInfo

		a.Unpack(absinfobits[:])
		dev.Absinfos[abscode] = a
		return a, nil
	}

	return a, ErrAbsBits

}

func (dev *Device) Sync() error {
	return syscall.Fsync(dev.fd)
}
