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

func ReadReplicaFromClient(c client.ReadReplicas, existingReplicas []ReadReplica, diags diag.Diagnostics) []ReadReplica {
	var res []ReadReplica
	for _, item := range c {
		// Convert parameter overrides
		paramOverrides := ParameterOverridesToMap(item.ParameterOverrides, diags)

		// Find matching replica in existing model to preserve null vs empty map
		if item.ParameterOverrides == nil || len(*item.ParameterOverrides) == 0 {
			// API returned empty - check if existing model had null
			for _, existingReplica := range existingReplicas {
				if existingReplica.Name.ValueString() == item.Name || existingReplica.ID.ValueString() == item.Id {
					if existingReplica.ParameterOverrides.IsNull() {
						paramOverrides = types.MapNull(types.StringType)
					}
					break
				}
			}
		}

		res = append(res, ReadReplica{
			Name:               types.StringValue(item.Name),
			ID:                 types.StringValue(item.Id),
			ParameterOverrides: paramOverrides,
		})
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

func ModelFromClient(postgres *client.PostgresDetail, connectionInfo *client.PostgresConnectionInfo, logStreamOverrides *logs.ResourceLogStreamSetting, existingModel PostgresModel, diags diag.Diagnostics) PostgresModel {
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
		ReadReplicas:            ReadReplicaFromClient(postgres.ReadReplicas, existingModel.ReadReplicas, diags),
		Version:                 types.StringValue(string(postgres.Version)),
		ConnectionInfo:          connectionInfoFromClient(connectionInfo, diags),
		LogStreamOverride:       common.LogStreamOverrideFromClient(logStreamOverrides, existingModel.LogStreamOverride, diags),
		DiskSizeGB:              common.IntPointerAsValue(postgres.DiskSizeGB),
		ParameterOverrides:      parameterOverrides,
	}
	return postgresModel
}
