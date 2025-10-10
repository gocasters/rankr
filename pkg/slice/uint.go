package slice

import types "github.com/gocasters/rankr/type"

func MapFromIDFloat64ToUint64Float64(m map[types.ID]float64) map[uint64]float64 {
	result := make(map[uint64]float64, len(m))
	for k, v := range m {
		result[uint64(k)] = v
	}
	return result
}

func MapFromUint64Float64ToIDFloat64(m map[uint64]float64) map[types.ID]float64 {
	result := make(map[types.ID]float64, len(m))
	for k, v := range m {
		result[types.ID(k)] = v
	}
	return result
}
