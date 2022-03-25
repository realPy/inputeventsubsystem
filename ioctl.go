package inputeventsubsystem

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

/*

#include <linux/input.h>
#include <stdlib.h>
#include <stdio.h>


static inline int eviocgname(int size)
{
return EVIOCGNAME(size);
}

static inline int eviocgphys(int size)
{
return EVIOCGPHYS(size);
}

static inline int eviocgbit(int min,int max)
{
return EVIOCGBIT(min,max);
}


static inline int eviocgabs(int type)
{
return EVIOCGABS(type);
}

static inline int eviockey(int size)
{
return EVIOCGKEY(size);
}


*/
import "C"

const (
	INPUT_NAME_LEN = 256
	INPUT_PHY_LEN  = 256
)

func ioctl(fd uintptr, name uintptr, data unsafe.Pointer) syscall.Errno {
	_, _, err := syscall.RawSyscall(syscall.SYS_IOCTL, fd, name, uintptr(data))
	return err
}

func IoctlInputName(fd int) (string, error) {
	var err error
	var value [INPUT_NAME_LEN]byte
	if errno := ioctl(uintptr(fd), uintptr(C.eviocgname(INPUT_NAME_LEN)), unsafe.Pointer(&value[0])); errno != 0 {
		err = errno
	}
	return unix.ByteSliceToString(value[:]), err

}

func IoctlInputPhys(fd int) (string, error) {
	var err error
	var value [INPUT_PHY_LEN]byte
	if errno := ioctl(uintptr(fd), uintptr(C.eviocgphys(INPUT_NAME_LEN)), unsafe.Pointer(&value[0])); errno != 0 {
		err = errno
	}

	return unix.ByteSliceToString(value[:]), err

}

func IoctlInputVersion(fd int) (uint32, error) {

	var version uint32
	var err error
	if errno := ioctl(uintptr(fd), C.EVIOCGVERSION, unsafe.Pointer(&version)); errno != 0 {
		err = errno
	}
	return version, err

}

func IoctlInputID(fd int) (uint16, uint16, uint16, uint16, error) {

	var ids [4]uint16

	var err error
	if errno := ioctl(uintptr(fd), C.EVIOCGID, unsafe.Pointer(&ids)); errno != 0 {
		err = errno
	}

	return ids[C.ID_BUS], ids[C.ID_VENDOR], ids[C.ID_PRODUCT], ids[C.ID_VERSION], err

}

func IoctlInputGrab(fd int, acquire bool) error {

	if acquire {
		return unix.IoctlSetInt(fd, C.EVIOCGRAB, 1)
	}

	return unix.IoctlSetInt(fd, C.EVIOCGRAB, 0)
}

func IoctlInputBit(fd int, min, max int) ([]byte, error) {

	var databits []byte = make([]byte, (max+1)/8)

	var err error
	if errno := ioctl(uintptr(fd), uintptr(C.eviocgbit(C.int(min), C.int(max))), unsafe.Pointer(&databits[0])); errno != 0 {
		err = errno
	}
	return databits, err
}

func IoctlInputAbs(fd int, typeabs int) ([]byte, error) {

	var absbits []byte = make([]byte, 24)

	var err error
	if errno := ioctl(uintptr(fd), uintptr(C.eviocgabs(C.int(typeabs))), unsafe.Pointer(&absbits[0])); errno != 0 {
		err = errno
	}
	return absbits, err
}

func IoctlInputKey(fd int) ([]byte, error) {

	var sizekeybits int = (KEY_MAX + 1) / 8

	var keybits []byte = make([]byte, sizekeybits)

	var err error
	if errno := ioctl(uintptr(fd), uintptr(C.eviockey(C.int(sizekeybits))), unsafe.Pointer(&keybits[0])); errno != 0 {
		err = errno
	}
	return keybits, err
}
