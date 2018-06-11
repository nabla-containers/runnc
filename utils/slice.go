package utils

// AddAbsentSlice adds an object to a slice of objects if it is not already present.
func AddAbsentSlice(slice []string, add string) []string {
	for _, v := range slice {
		if add == v {
			return slice
		}
	}

	return append(slice, add)
}
