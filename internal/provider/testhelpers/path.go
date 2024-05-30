package testhelpers

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func ExamplesPath(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err)

	index := strings.LastIndex(dir, "terraform-provider")
	if index == -1 {
		t.Fatalf("could not find 'terraform-provider' in the path")
	}

	parentDir := dir[:index+len("terraform-provider")]
	dir = path.Join(parentDir, "examples")
	return dir
}
