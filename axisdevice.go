package inputeventsubsystem

import (
	"errors"
	"math"
)

var (
	ErrDeviceHasNoAxis = errors.New("the device has no axis")
)

type axis_correct struct {
	Coef    []int32
	Maximum int32
	Minimum int32
	Scale   float32
}

type AxisDevice struct {
	Device           *Device
	useDeadZone      bool
	axis_corrections map[int]axis_correct
}

func (e *Device) AxisDevice(UseDeadZone bool, overrideFlatValue int) (*AxisDevice, error) {
	var ad AxisDevice
	var err error

	if _, ok := e.Capabilities[EV_ABS]; ok {

		ad.axis_corrections = make(map[int]axis_correct, 0)
		ad.useDeadZone = UseDeadZone
		ad.Device = e

		for abs_code, absinfo := range e.Absinfos {
			var a axis_correct

			a.Coef = make([]int32, 3)
			if UseDeadZone {
				a.Maximum = absinfo.Maximum
				a.Minimum = absinfo.Minimum

				if overrideFlatValue >= 0 {
					absinfo.Flat = int32(overrideFlatValue)
				}

				a.Coef[0] = (absinfo.Maximum + absinfo.Minimum) - 2*absinfo.Flat

				a.Coef[1] = (absinfo.Maximum + absinfo.Minimum) + 2*absinfo.Flat
				t := ((absinfo.Maximum - absinfo.Minimum) - 4*absinfo.Flat)

				if t != 0 {
					a.Coef[2] = (1 << 28) / t
				} else {
					a.Coef[2] = 0
				}
			} else {

				var value_range float64 = float64(absinfo.Maximum - absinfo.Minimum)
				var output_range float64 = float64(32767 * 2)

				a.Scale = float32(output_range / value_range)

			}

			ad.axis_corrections[abs_code] = a

		}
	} else {
		err = ErrDeviceHasNoAxis
	}

	return &ad, err
}

func (a *AxisDevice) CorrectAxis(which int, value int32) int32 {

	if a.useDeadZone {

		if array_correction, ok := a.axis_corrections[which]; !ok {

			return value
		} else {
			value = value * 2
			if value > array_correction.Coef[0] {
				if value < array_correction.Coef[1] {
					return 0
				}

				value = value - array_correction.Coef[1]
			} else {
				value = value - array_correction.Coef[0]
			}
			value = value * array_correction.Coef[2]
			value >>= 13
		}

	} else {
		if array_correction, ok := a.axis_corrections[which]; !ok {

			return value
		} else {
			value = int32(math.Floor(float64(float32(value-array_correction.Minimum)*array_correction.Scale - float32(32768) + 0.5)))
		}
	}

	/* Clamp and return */
	if value < -32767 {
		return -32767
	}
	if value > 32767 {
		return 32767
	}
	return value

}

func (a *AxisDevice) Read() chan []*Event {

	return a.Device.Read()

}

func (a *AxisDevice) ReadDone(events []*Event) {

	a.Device.ReadDone(events)

}

func (a *AxisDevice) Close() error {
	return a.Device.Close()
}
