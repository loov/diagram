package sequence

import "math"

type Time = float64

var Automatic Time = math.NaN()

func IsAutomatic(t Time) bool { return math.IsNaN(t) }
func Min(a, b Time) Time {
	if IsAutomatic(a) {
		return b
	}
	if a < b {
		return a
	}
	return b
}
func Max(a, b Time) Time {
	if IsAutomatic(a) {
		return b
	}
	if a > b {
		return a
	}
	return b
}
