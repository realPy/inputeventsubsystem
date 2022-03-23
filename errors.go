package inputeventsubsystem

import "errors"

var (
	ErrDriverVersion     = errors.New("unable to get driver version")
	ErrDeviceInformation = errors.New("unable to get device information")
	ErrAbsBits           = errors.New("unable to get absbits")
	ErrEvBits            = errors.New("unable to get evbits")
)
