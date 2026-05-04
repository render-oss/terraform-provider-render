package postgres

import (
	"fmt"

	"terraform-provider-render/internal/client/logs"
	"terraform-provider-render/internal/provider/common"
	commontypes "terraform-provider-render/internal/provider/common/types"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
)

type PostgresModel struct {
	DatadogAPIKey           types.String                  `tfsdk:"datadog_api_key"`
	DatabaseName            commontypes.SuffixStringValue `tfsdk:"database_name"`
	DatabaseUser            types.String                  `tfsdk:"database_user"`
	EnvironmentID           types.String                  `tfsdk:"environment_id"`
	HighAvailabilityEnabled types.Bool                    `tfsdk:"high_availability_enabled"`
	ID                      types.String                  `tfsdk:"id"`
	IPAllowList             types.Set                     `tfsdk:"ip_allow_list"`
	Name                    types.String                  `tfsdk:"name"`
	Plan                    types.String                  `tfsdk:"plan"`
	PrimaryPostgresID       types.String                  `tfsdk:"primary_postgres_id"`
	ReadReplicas            []ReadReplica                 `tfsdk:"read_replicas"`
	Region                  types.String                  `tfsdk:"region"`
	Role                    types.String                  `tfsdk:"role"`
	Version                 types.String                  `tfsdk:"version"`
	ConnectionInfo          types.Object                  `tfsdk:"connection_info"`
	LogStreamOverride       types.Object                  `tfsdk:"log_stream_override"`
	DiskSizeGB              types.Int64                   `tfsdk:"disk_size_gb"`
	ParameterOverrides      types.Map                     `tfsdk:"parameter_overrides"`
}

type ReadReplica struct {
	Name               types.String `tfsdk:"name"`
	ID                 types.String `tfsdk:"id"`
	ParameterOverrides types.Map    `tfsdk:"parameter_overrides"`
	LogStreamOverride  types.Object `tfsdk:"log_stream_override"`
}

type ConnectionInfo struct {
	Password                 types.String `tfsdk:"password"`
	ExternalConnectionString types.String `tfsdk:"external_connection_string"`
	InternalConnectionString types.String `tfsdk:"internal_connection_string"`
	PSQLCommand              types.String `tfsdk:"psql_command"`
}

// ParameterOverridesToMap converts API parameter overrides to Terraform map
func ParameterOverridesToMap(po *client.PostgresParameterOverrides, diags diag.Diagnostics) types.Map {
	elements := make(map[string]attr.Value)

	if po != nil {
		for k, v := range *po {
			elements[k] = types.StringValue(v)
		}
	}

	mapValue, mapDiags := types.MapValue(types.StringType, elements)
	diags.Append(mapDiags...)

	return mapValue
}

// ParameterOverridesToGoMap converts Terraform map to Go map for API requests
func ParameterOverridesToGoMap(m types.Map, diags diag.Diagnostics) *client.PostgresParameterOverrides {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}

	elements := m.Elements()
	if len(elements) == 0 {
		// Empty map means clear all overrides
		emptyMap := client.PostgresParameterOverrides{}
		return &emptyMap
	}

	goMap := make(client.PostgresParameterOverrides)
	for k, v := range elements {
		strVal, ok := v.(types.String)
		if !ok {
			diags.AddError(
				"Invalid Parameter Override Type",
				fmt.Sprintf("Parameter override '%s' expected string value, got %T. This indicates a schema configuration error.", k, v),
			)
			continue
		}
		goMap[k] = strVal.ValueString()
	}

	return &goMap
}

func ReadReplicaFromClient(c client.ReadReplicas, existingReplicas []ReadReplica, replicaLogStreams map[string]*logs.ResourceLogStreamSetting, diags diag.Diagnostics) []ReadReplica {
	// Index API replicas by name. read_replicas is a List on the TF side, but
	// the API returns replicas in unspecified SQL order (see PGClusterReplicas
	// in the api repo's pkg/models/postgresdb.go — no ORDER BY). Sorting the
	// returned slice to match existingReplicas order keeps state aligned with
	// the user's HCL order regardless of how the API sorts the response.

	// Preserve the plan's null-vs-empty distinction. With ListNestedAttribute,
	// TF treats null and [] as different; returning [] when the plan was null
	// (or vice versa) produces an "inconsistent result after apply" error.
	// existingReplicas reflects what the user wrote in HCL: nil for omitted /
	// `= null`, non-nil-empty for an explicit `= []`.
	if len(c) == 0 {
		if existingReplicas == nil {
			return nil
		}
		return []ReadReplica{}
	}

	apiByName := make(map[string]client.ReadReplica, len(c))
	for _, item := range c {
		apiByName[item.Name] = item
	}

	res := make([]ReadReplica, 0, len(c))
	consumed := make(map[string]struct{}, len(c))

	build := func(item client.ReadReplica, existing *ReadReplica) ReadReplica {
		paramOverrides := ParameterOverridesToMap(item.ParameterOverrides, diags)
		if item.ParameterOverrides == nil || len(*item.ParameterOverrides) == 0 {
			if existing != nil && existing.ParameterOverrides.IsNull() {
				paramOverrides = types.MapNull(types.StringType)
			}
		}

		var existingLSO types.Object
		if existing != nil {
			existingLSO = existing.LogStreamOverride
		}

		return ReadReplica{
			Name:               types.StringValue(item.Name),
			ID:                 types.StringValue(item.Id),
			ParameterOverrides: paramOverrides,
			LogStreamOverride:  common.LogStreamOverrideFromClient(replicaLogStreams[item.Id], existingLSO, diags),
		}
	}

	// First pass: emit replicas in existingReplicas order, matched by name.
	// This preserves the user's HCL ordering across refresh cycles in the
	// resource path (where existingReplicas is plan/state). It also makes
	// the no-op for the datasource path, where existingReplicas is always
	// nil because read_replicas is Computed-only there.
	for i := range existingReplicas {
		existing := &existingReplicas[i]
		name := existing.Name.ValueString()
		if name == "" {
			continue
		}
		item, ok := apiByName[name]
		if !ok {
			continue
		}
		res = append(res, build(item, existing))
		consumed[name] = struct{}{}
	}

	// Second pass: append any API replicas the first pass didn't match.
	// In the datasource this is the only branch used (existingReplicas is
	// nil → first pass emits nothing). In the resource it's reached only
	// when the API returns a replica the existing model doesn't know about
	// — e.g. drift introduced out-of-band. Calling build with existing=nil
	// means token will be types.StringNull, which matches the datasource's
	// established behavior on the top-level log_stream_override (the API
	// doesn't return tokens, so we have nothing to surface).
	for _, item := range c {
		if _, done := consumed[item.Name]; done {
			continue
		}
		res = append(res, build(item, nil))
	}

	return res
}

func ReadReplicaInputFromModel(r []ReadReplica, diags diag.Diagnostics) []client.ReadReplicaInput {
	var res []client.ReadReplicaInput
	for _, item := range r {
		res = append(res, client.ReadReplicaInput{
			Name:               item.Name.ValueString(),
			ParameterOverrides: ParameterOverridesToGoMap(item.ParameterOverrides, diags),
		})
	}
	return res
}

var connectionInfoTypes = map[string]attr.Type{
	"password":                   types.StringType,
	"external_connection_string": types.StringType,
	"internal_connection_string": types.StringType,
	"psql_command":               types.StringType,
}

func connectionInfoFromClient(c *client.PostgresConnectionInfo, diags diag.Diagnostics) types.Object {
	if c == nil {
		return types.ObjectNull(connectionInfoTypes)
	}

	objectValue, objectDiags := types.ObjectValue(
		connectionInfoTypes,
		map[string]attr.Value{
			"password":                   types.StringValue(c.Password),
			"external_connection_string": types.StringValue(c.ExternalConnectionString),
			"internal_connection_string": types.StringValue(c.InternalConnectionString),
			"psql_command":               types.StringValue(c.PsqlCommand),
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func ModelFromClient(postgres *client.PostgresDetail, connectionInfo *client.PostgresConnectionInfo, logStreamOverrides *logs.ResourceLogStreamSetting, replicaLogStreams map[string]*logs.ResourceLogStreamSetting, existingModel PostgresModel, diags diag.Diagnostics) PostgresModel {
	// Handle parameter_overrides: preserve null if it was null in existing model
	parameterOverrides := ParameterOverridesToMap(postgres.ParameterOverrides, diags)
	if existingModel.ParameterOverrides.IsNull() && (postgres.ParameterOverrides == nil || len(*postgres.ParameterOverrides) == 0) {
		// Existing model had null and API returned empty/null -> keep as null
		parameterOverrides = types.MapNull(types.StringType)
	}

	postgresModel := PostgresModel{
		ID:                      types.StringValue(postgres.Id),
		Name:                    types.StringValue(postgres.Name),
		IPAllowList:             common.IPAllowListFromClient(postgres.IpAllowList, diags),
		DatadogAPIKey:           existingModel.DatadogAPIKey,
		DatabaseName:            commontypes.SuffixStringValue{StringValue: types.StringValue(postgres.DatabaseName)},
		DatabaseUser:            types.StringValue(postgres.DatabaseUser),
		EnvironmentID:           types.StringPointerValue(postgres.EnvironmentId),
		Plan:                    types.StringValue(string(postgres.Plan)),
		PrimaryPostgresID:       types.StringPointerValue(postgres.PrimaryPostgresID),
		Region:                  types.StringValue(string(postgres.Region)),
		Role:                    types.StringValue(string(postgres.Role)),
		HighAvailabilityEnabled: types.BoolValue(postgres.HighAvailabilityEnabled),
		ReadReplicas:            ReadReplicaFromClient(postgres.ReadReplicas, existingModel.ReadReplicas, replicaLogStreams, diags),
		Version:                 types.StringValue(string(postgres.Version)),
		ConnectionInfo:          connectionInfoFromClient(connectionInfo, diags),
		LogStreamOverride:       common.LogStreamOverrideFromClient(logStreamOverrides, existingModel.LogStreamOverride, diags),
		DiskSizeGB:              common.IntPointerAsValue(postgres.DiskSizeGB),
		ParameterOverrides:      parameterOverrides,
	}
	return postgresModel
}
