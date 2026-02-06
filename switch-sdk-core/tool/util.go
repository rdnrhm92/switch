package tool

func ConvertInt32Array2IntArray(array []int32) []int {
	ints := make([]int, len(array))
	for i, v := range array {
		ints[i] = int(v)
	}
	return ints
}

func UniqueSlice(slice []uint) []uint {
	if len(slice) == 0 {
		return slice
	}

	keys := make(map[uint]struct{}, len(slice))
	list := make([]uint, 0, len(slice))

	for _, entry := range slice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}
	return list
}
