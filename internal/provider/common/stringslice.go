package common

import (
	"regexp"
	"slices"
	"strings"
)

// EnumStringsExcept converts generated enum values, such as
// client.RedisPlanValues(), to plain strings, dropping the excluded ones. It
// lets validators and descriptions derive their values from the OpenAPI schema
// instead of a hand-maintained list. Exclusions cover enum members that are not
// valid inputs on their own, such as the "custom" plan, whose real name is matched
// by pattern instead.
func EnumStringsExcept[T ~string](values []T, exclude ...T) []string {
	strs := make([]string, 0, len(values))
	for _, v := range values {
		if !slices.Contains(exclude, v) {
			strs = append(strs, string(v))
		}
	}
	return strs
}

// EnumValuesMatching returns the enum values matching re. Descriptions use it
// to list one naming scheme's values while covering the rest as prose.
func EnumValuesMatching[T ~string](values []T, re *regexp.Regexp) []T {
	matched := make([]T, 0, len(values))
	for _, v := range values {
		if re.MatchString(string(v)) {
			matched = append(matched, v)
		}
	}
	return matched
}

// EnumList renders enum values as a comma-separated list for an attribute's
// Description, such as "free, starter, standard".
func EnumList[T ~string](values []T, exclude ...T) string {
	return enumList(values, "", exclude)
}

// EnumListMarkdown is EnumList with each value in backticks, for an attribute's
// MarkdownDescription.
func EnumListMarkdown[T ~string](values []T, exclude ...T) string {
	return enumList(values, "`", exclude)
}

func enumList[T ~string](values []T, quote string, exclude []T) string {
	strs := EnumStringsExcept(values, exclude...)
	for i, s := range strs {
		strs[i] = quote + s + quote
	}

	return strings.Join(strs, ", ")
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
