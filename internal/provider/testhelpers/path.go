package testhelpers

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const parentDirectory = "terraform-provider-render"

func ExamplesPath(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err)

	index := strings.LastIndex(dir, parentDirectory)
	if index == -1 {
		t.Fatalf("could not find '%s' in the path", parentDirectory)
	}

	parentDir := dir[:index+len(parentDirectory)]
	dir = path.Join(parentDir, "examples")
	return dir
}
