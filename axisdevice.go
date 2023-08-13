package inputeventsubsystem

import (
	"errors"
)

var (
	ErrDeviceHasNoAxis = errors.New("the device has no axis")
)

type axis_correct struct {
	Coef    []int32
	Maximum int32
	Minimum int32
	Median  int32
	Scale   float64
}

type AxisDevice struct {
	Device           *Device
	useDeadZone      bool
	axis_corrections map[int]axis_correct
}

func CreateAxisDeviceFromAbsInfo(e *Device, absinfos map[int]AbsInfo, useDeadZone bool, overrideFlatValue int) *AxisDevice {
	var ad AxisDevice

	ad.axis_corrections = make(map[int]axis_correct, 0)
	ad.useDeadZone = useDeadZone
	ad.Device = e

	for abs_code, absinfo := range absinfos {
		var a axis_correct

		a.Coef = make([]int32, 3)

		if absinfo.Maximum == 0 && absinfo.Minimum == 0 {
			absinfo.Minimum = -32767
			absinfo.Maximum = 32767
		}

		if useDeadZone {

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
		}

		a.Maximum = absinfo.Maximum
		a.Minimum = absinfo.Minimum

		var value_range float64 = float64(absinfo.Maximum - absinfo.Minimum - 1)
		var output_range float64 = float64(65534)

		a.Scale = float64(output_range) / float64(value_range)
		a.Median = int32(value_range) / 2

		ad.axis_corrections[abs_code] = a

	}

	return &ad
}

func (e *Device) AxisDevice(UseDeadZone bool, overrideFlatValue int) (*AxisDevice, error) {
	var ad *AxisDevice
	var err error

	if _, ok := e.Capabilities[EV_ABS]; ok {

		ad = CreateAxisDeviceFromAbsInfo(e, e.Absinfos, UseDeadZone, overrideFlatValue)

	} else {
		err = ErrDeviceHasNoAxis
	}

	return ad, err
}

func (a *AxisDevice) GetRange(which int) (int32, int32) {
	if array_correction, ok := a.axis_corrections[which]; ok {
		return array_correction.Minimum, array_correction.Maximum
	}
	return 0, 0
}

func (a *AxisDevice) CorrectAxisWithDeadzone(which int, value int32, deadzone bool) int32 {

	if a.useDeadZone && deadzone {

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
			offset_scale := (value - array_correction.Minimum)

			value = int32(float64(offset_scale-array_correction.Median) * array_correction.Scale)

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

func (a *AxisDevice) GetRange(which int) (int32, int32) {
	if array_correction, ok := a.axis_corrections[which]; !ok {
		return array_correction.Minimum, array_correction.Maximum
	}
	return 0, 0
}

func (a *AxisDevice) CorrectAxis(which int, value int32) int32 {

	return a.CorrectAxisWithDeadzone(which, value, true)

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
