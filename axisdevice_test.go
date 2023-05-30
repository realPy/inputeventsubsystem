package inputeventsubsystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrectAxis(t *testing.T) {
	t.Run("0-255 range deadzone 0", func(t *testing.T) {
		a := CreateAxisDeviceFromAbsInfo(nil, map[int]AbsInfo{ABS_X: {Minimum: 0, Maximum: 255, Flat: 0}}, true, -1)

		v := a.CorrectAxis(ABS_X, 100)
		assert.Equal(t, int32(-7068), v)
	})

	t.Run("0-255 range deadzone override 10", func(t *testing.T) {
		a := CreateAxisDeviceFromAbsInfo(nil, map[int]AbsInfo{ABS_X: {Minimum: 0, Maximum: 255, Flat: 0}}, true, 10)

		v := a.CorrectAxis(ABS_X, 10)
		assert.Equal(t, int32(-32767), v)
	})

	t.Run("-2 - 2  range deadzone override 1", func(t *testing.T) {
		a := CreateAxisDeviceFromAbsInfo(nil, map[int]AbsInfo{ABS_X: {Minimum: -2, Maximum: 2, Flat: 0}}, true, 1)

		v := a.CorrectAxis(ABS_X, 10)
		assert.Equal(t, int32(0), v)
	})

	t.Run("0-0 range deadzone 0", func(t *testing.T) {
		a := CreateAxisDeviceFromAbsInfo(nil, map[int]AbsInfo{ABS_X: {Minimum: 0, Maximum: 0, Flat: 0}}, true, -1)

		v := a.CorrectAxis(ABS_X, 1000)
		assert.Equal(t, int32(1000), v)
	})

	t.Run("0-255 range deadzone off", func(t *testing.T) {
		a := CreateAxisDeviceFromAbsInfo(nil, map[int]AbsInfo{ABS_X: {Minimum: 0, Maximum: 0, Flat: 0}}, false, -1)

		v := a.CorrectAxis(ABS_X, 100)
		assert.Equal(t, int32(-32668), v)
	})

	t.Run("0-255 range deadzone override 10 divide 255", func(t *testing.T) {
		a := CreateAxisDeviceFromAbsInfo(nil, map[int]AbsInfo{ABS_X: {Minimum: 0, Maximum: 255, Flat: 0}}, true, 0)
		v := a.CorrectAxis(ABS_X, 100)/255 + 127
		assert.Equal(t, int32(100), v)
	})

}
