package common_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"terraform-provider-render/internal/provider/common"
)

func Test_XORStringSlices(t *testing.T) {
	t.Parallel()
	t.Run("Same slices return no results", func(t *testing.T) {
		t.Parallel()
		slice1 := []string{"a", "b", "c"}
		slice2 := []string{"c", "b", "a"}
		result1, both, result2 := common.XORStringSlices(slice1, slice2)
		if len(result1) != 0 || len(result2) != 0 {
			t.Errorf("Expected no results, got %v, %v", result1, result2)
		}

		if len(both) != 3 {
			t.Errorf("Expected 3 results, got %v", both)
		}
	})

	t.Run("Slices with no elements in common return all values", func(t *testing.T) {
		t.Parallel()
		slice1 := []string{"a", "b", "c"}
		slice2 := []string{"d", "e", "f"}
		result1, both, result2 := common.XORStringSlices(slice1, slice2)
		if len(result1) != 3 || len(result2) != 3 {
			t.Errorf("Expected 3 results, got %v, %v", result1, result2)
		}
		if len(both) != 0 {
			t.Errorf("Expected no results, got %v", both)
		}
	})

	t.Run("Slices with some elements in common return only unique values", func(t *testing.T) {
		t.Parallel()
		slice1 := []string{"a", "b", "c"}
		slice2 := []string{"b", "c", "d"}
		result1, both, result2 := common.XORStringSlices(slice1, slice2)
		require.Equal(t, []string{"a"}, result1)
		require.Equal(t, []string{"d"}, result2)
		require.Equal(t, []string{"b", "c"}, both)
	})
}

func Test_EnumList(t *testing.T) {
	t.Parallel()

	type plan string
	values := []plan{"free", "starter", "custom", "1g"}

	t.Run("renders a comma-separated list", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "free, starter, custom, 1g", common.EnumList(values))
	})

	t.Run("drops excluded values", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "free, starter, 1g", common.EnumList(values, "custom"))
	})

	t.Run("quotes values for markdown", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "`free`, `starter`, `1g`", common.EnumListMarkdown(values, "custom"))
	})

	t.Run("empty values render an empty list", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "", common.EnumList([]plan{}))
	})
}

func Test_EnumValuesMatching(t *testing.T) {
	t.Parallel()

	type plan string
	values := []plan{"free", "basic_256mb", "1c-4g", "pro_4gb", "128c-1024g"}
	matched := common.EnumValuesMatching(values, regexp.MustCompile(`^[\d.]+c-\d+(mb|g)$`))

	require.Equal(t, []plan{"1c-4g", "128c-1024g"}, matched)
}
