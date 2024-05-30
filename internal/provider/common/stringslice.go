package common

// XORStringSlices returns two slices, one with elements that are in slice1 but not in slice2, and the other with elements that are in slice2 but not in slice1.
func XORStringSlices(slice1, slice2 []string) (inFirst []string, inBoth []string, inSecond []string) {
	map1 := make(map[string]bool)
	map2 := make(map[string]bool)

	for _, item := range slice1 {
		map1[item] = true
	}

	for _, item := range slice2 {
		map2[item] = true
	}

	for _, item := range slice1 {
		if !map2[item] {
			inFirst = append(inFirst, item)
		}
	}

	for _, item := range slice1 {
		if map2[item] {
			inBoth = append(inBoth, item)
		}
	}

	for _, item := range slice2 {
		if !map1[item] {
			inSecond = append(inSecond, item)
		}
	}

	return inFirst, inBoth, inSecond
}
