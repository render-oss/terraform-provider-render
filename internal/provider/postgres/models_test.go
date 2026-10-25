package postgres

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
)

func TestReadReplicaFromClientPreservesEmptySet(t *testing.T) {
	tests := []struct {
		name     string
		existing []ReadReplica
		api      client.ReadReplicas
		wantNil  bool
	}{
		{name: "null configuration", wantNil: true},
		{name: "explicit empty configuration", existing: []ReadReplica{}, api: client.ReadReplicas{}},
		{name: "replicas removed remotely", existing: []ReadReplica{{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadReplicaFromClient(tt.api, tt.existing, nil, diag.Diagnostics{})
			require.Empty(t, got)
			assert.Equal(t, tt.wantNil, got == nil)
		})
	}
}
