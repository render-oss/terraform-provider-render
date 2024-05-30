package common_test

import (
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
