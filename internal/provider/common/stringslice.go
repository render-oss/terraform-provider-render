package common

import "slices"

// EnumStrings converts generated enum values, such as client.RedisPlanValues(),
// to plain strings. It lets validators derive their allowed values from the
// OpenAPI schema instead of a hand-maintained list that goes stale whenever the
// schema gains a value.
func EnumStrings[T ~string](values []T) []string {
	return EnumStringsExcept(values)
}

// EnumStringsExcept is EnumStrings without the excluded values. Plan validators
// use it to drop enum members that are not valid inputs on their own, such as
// the "custom" plan, whose real name is matched by pattern instead.
func EnumStringsExcept[T ~string](values []T, exclude ...T) []string {
	strs := make([]string, 0, len(values))
	for _, v := range values {
		if !slices.Contains(exclude, v) {
			strs = append(strs, string(v))
		}
	}
	return strs
}

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
