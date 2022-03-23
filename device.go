package inputeventsubsystem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

/*

#include <linux/input.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <errno.h>
#include <fcntl.h>
#include <string.h>
#include <pthread.h>
#include <sched.h>

static inline int getdevicename(int fd,int size,char *name_dst)
{
return ioctl(fd, EVIOCGNAME(size), name_dst);
}

static inline int getphy(int fd,int size,char *phy_dst)
{
return ioctl(fd, EVIOCGPHYS(size), phy_dst);
}

static inline int geteviocbits(int fd,int min,int max,void *evbits)
{
return ioctl(fd,EVIOCGBIT(min, max),evbits);
}

static inline int geteviocgkey(int fd,int size,void *keybits)
{
return ioctl(fd,EVIOCGKEY(size),keybits);
}

static inline int geteviocgabs(int fd,int type,void *absbits)
{
return ioctl(fd,EVIOCGABS(type),absbits);
}


*/
import "C"

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
	pipe          unsafe.Pointer
}

func (e *Device) String() string {
	return fmt.Sprintf("%s: bus 0x%x vendor 0x%x product 0x%x version 0x%x\n", e.Name, e.bus, e.VendorID, e.ProductID, e.VendorID)
}

// Open an evdev input device.
func Open(devnode string) (*Device, error) {

	var dev Device
	dev.Fn = devnode

	f, err := syscall.Open(dev.Fn, syscall.O_RDONLY, 0666)

	if err != nil {
		return nil, err
	}

	dev.fd = f

	ierr := ioctl(uintptr(dev.fd), C.EVIOCGVERSION, unsafe.Pointer(&dev.DriverVersion))

	if ierr != 0 {
		dev.DriverVersion = 0
		defer syscall.Close(dev.fd)
		return nil, ErrDriverVersion

	}

	ids := new([4]uint16)

	ierr = ioctl(uintptr(dev.fd), C.EVIOCGID, unsafe.Pointer(ids))

	if ierr != 0 {
		defer syscall.Close(dev.fd)
		return nil, ErrDeviceInformation
	}

	dev.ProductID = ids[C.ID_PRODUCT]
	dev.VendorID = ids[C.ID_VENDOR]
	dev.bus = ids[C.ID_BUS]
	dev.Version = ids[C.ID_VERSION]
	dev.eventchan = make(chan []*Event, 1)

	ptrname := C.malloc(C.sizeof_char * 256)

	defer C.free(unsafe.Pointer(ptrname))

	C.getdevicename(C.int(dev.fd), 256, (*C.char)(ptrname))
	dev.Name = C.GoString((*C.char)(ptrname))

	ptrphy := C.malloc(C.sizeof_char * 256)

	defer C.free(unsafe.Pointer(ptrphy))

	C.getphy(C.int(dev.fd), 256, (*C.char)(ptrphy))
	dev.Phy = C.GoString((*C.char)(ptrphy))

	evbits := new([(EV_MAX + 1) / 8]byte)

	if errevbits := C.geteviocbits(C.int(dev.fd), 0, EV_MAX, unsafe.Pointer(evbits)); errevbits < 0 {
		defer syscall.Close(dev.fd)
		return nil, errors.New("unable to get evbits")
	}

	dev.Capabilities = make(map[int]map[int]string)
	dev.Absinfos = make(map[int]AbsInfo)

	for evtype := 0; evtype < EV_MAX; evtype++ {
		if evbits[evtype/8]&(1<<uint(evtype%8)) != 0 {
			dev.Capabilities[evtype] = make(map[int]string)
			if evtype == EV_KEY {
				codebits := new([(KEY_MAX + 1) / 8]byte)
				if errevbits := C.geteviocbits(C.int(dev.fd), C.int(evtype), KEY_MAX, unsafe.Pointer(codebits)); errevbits >= 0 {

					for evcode := 0; evcode < KEY_MAX; evcode++ {
						if codebits[evcode/8]&(1<<uint(evcode%8)) != 0 {
							dev.Capabilities[evtype][evcode] = fmt.Sprintf("0x%x", evcode)

						}
					}

				}

			}

			if evtype == EV_ABS {

				absbits := new([(ABS_MAX + 1) / 8]byte)

				if errevbits := C.geteviocbits(C.int(dev.fd), C.int(evtype), KEY_MAX, unsafe.Pointer(absbits)); errevbits >= 0 {

					for abscode := 0; abscode < ABS_MAX; abscode++ {
						if absbits[abscode/8]&(1<<uint(abscode%8)) != 0 {

							dev.Capabilities[evtype][abscode] = fmt.Sprintf("0x%x", abscode)

							//hat not have absinfo
							if abscode < ABS_HAT0X || abscode > ABS_HAT3Y {
								absinfobits := new([24]byte)

								if err := C.geteviocgabs(C.int(dev.fd), C.int(abscode), unsafe.Pointer(absinfobits)); err >= 0 {
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

func (dev *Device) Read() chan []*Event {

	go func() {
		var events [deviceinputeventsize * 64]byte

		for {

			if n, err := syscall.Read(dev.fd, events[:]); err == nil {
				p := UnpackDeviceInputEvents(events[0:n])
				dev.eventchan <- p

			} else {

				return
			}

		}

	}()
	return dev.eventchan
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
	return syscall.Close(dev.fd)
}

func (dev *Device) KeysState() ([]byte, error) {
	var sizekeybits int = (KEY_MAX + 1) / 8

	keybits := new([(KEY_MAX + 1) / 8]byte)

	if errkeybits := C.geteviocgkey(C.int(dev.fd), C.int(sizekeybits), unsafe.Pointer(keybits)); errkeybits < 0 {

		return nil, ErrEvBits
	}
	return keybits[:], nil
}

func (dev *Device) AbsState(abscode int) (AbsInfo, error) {
	var a AbsInfo

	absbits := new([(ABS_MAX + 1) / 8]byte)

	if errabsbits := C.geteviocgabs(C.int(dev.fd), C.int(abscode), unsafe.Pointer(absbits)); errabsbits < 0 {

		return a, ErrAbsBits
	}

	a.Unpack(absbits[:])
	return a, nil
}

func (dev *Device) Sync() error {
	return syscall.Fsync(dev.fd)
}
