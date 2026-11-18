package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"terraform-provider-render/internal/client"
)

func TestShouldDeployAfterUpdate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		explicitSkip           bool
		postUpdateServiceState client.ServiceSuspended
		want                   bool
	}{
		"explicit skip with active service": {
			explicitSkip:           true,
			postUpdateServiceState: client.ServiceSuspendedNotSuspended,
			want:                   false,
		},
		"explicit skip with suspended service": {
			explicitSkip:           true,
			postUpdateServiceState: client.ServiceSuspendedSuspended,
			want:                   false,
		},
		"suspended service": {
			postUpdateServiceState: client.ServiceSuspendedSuspended,
			want:                   false,
		},
		"active service": {
			postUpdateServiceState: client.ServiceSuspendedNotSuspended,
			want:                   true,
		},
		"unknown state preserves deployment": {
			postUpdateServiceState: client.ServiceSuspended("future-state"),
			want:                   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, shouldDeployAfterUpdate(test.explicitSkip, test.postUpdateServiceState))
		})
	}
}
