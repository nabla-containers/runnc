package utils

import "math"

func ConvertBytes(value *int64) int64 {
	mbSize := int64(float64(*value) / math.Pow(float64(1024), float64(2)))
	return mbSize
}
