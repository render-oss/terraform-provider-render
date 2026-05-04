package postgres

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/logs"
	"terraform-provider-render/internal/provider/common"
)

// GetReplicaLogStreamOverrides fetches the log stream override for each replica
// in pgReplicas. Returns a map keyed by replica ID. Replicas with no override
// are omitted from the map.
func GetReplicaLogStreamOverrides(ctx context.Context, apiClient *client.ClientWithResponses, pgReplicas client.ReadReplicas) (map[string]*logs.ResourceLogStreamSetting, error) {
	out := make(map[string]*logs.ResourceLogStreamSetting, len(pgReplicas))
	for _, replica := range pgReplicas {
		replicaLSO, err := common.GetLogStreamOverrides(ctx, apiClient, replica.Id)
		if err != nil {
			return nil, fmt.Errorf("replica %s: %w", replica.Id, err)
		}
		if replicaLSO != nil {
			out[replica.Id] = replicaLSO
		}
	}
	return out, nil
}

// UpdateReplicaLogStreamOverrides applies the per-replica log_stream_override
// changes described by plan and state. pgReplicas is the post-PATCH/POST replica
// list returned by the API — the source of truth for replica IDs. plan and state
// are the TF-level replica slices and are matched to pgReplicas by name; name is
// Required in the schema and unique within a primary, so it's a stable join key.
//
// Returns a map keyed by replica ID with the resulting log stream setting.
// Replicas whose update produced no setting are omitted.
//
// Edge case (intentionally not handled here): if a user renames a replica in
// HCL (state name=A, plan name=B), the API treats this as delete-A + create-B
// at the parent PATCH (see partitionReplicaNames in pkg/userdb/databaseservice/
// apiservice.go). The deleted replica's log_stream_setting row is cleaned up
// server-side as part of the postgres deletion path; we do not need to issue
// an explicit DELETE here. If the API ever moves to in-place rename (no
// recreate), this would silently leak the old override.
func UpdateReplicaLogStreamOverrides(ctx context.Context, apiClient *client.ClientWithResponses, pgReplicas client.ReadReplicas, plan, state []ReadReplica) (map[string]*logs.ResourceLogStreamSetting, error) {
	planByName := lsoByReplicaName(plan)
	stateByName := lsoByReplicaName(state)

	out := make(map[string]*logs.ResourceLogStreamSetting, len(pgReplicas))
	for _, pgReplica := range pgReplicas {
		replicaLSO, err := common.UpdateLogStreamOverride(
			ctx,
			apiClient,
			pgReplica.Id,
			&common.LogStreamOverrideStateAndPlan{
				Plan:  planByName[pgReplica.Name],
				State: stateByName[pgReplica.Name],
			},
		)
		if err != nil {
			return nil, fmt.Errorf("replica %s: %w", pgReplica.Id, err)
		}
		if replicaLSO != nil {
			out[pgReplica.Id] = replicaLSO
		}
	}
	return out, nil
}

func lsoByReplicaName(replicas []ReadReplica) map[string]types.Object {
	m := make(map[string]types.Object, len(replicas))
	for _, replica := range replicas {
		m[replica.Name.ValueString()] = replica.LogStreamOverride
	}
	return m
}
