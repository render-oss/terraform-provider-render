package postgres

import (
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
}

type ReadReplica struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
}

type ConnectionInfo struct {
	Password                 types.String `tfsdk:"password"`
	ExternalConnectionString types.String `tfsdk:"external_connection_string"`
	InternalConnectionString types.String `tfsdk:"internal_connection_string"`
	PSQLCommand              types.String `tfsdk:"psql_command"`
}

func ReadReplicaFromClient(c client.ReadReplicas) []ReadReplica {
	var res []ReadReplica
	for _, item := range c {
		res = append(res, ReadReplica{
			Name: types.StringValue(item.Name),
			ID:   types.StringValue(item.Id),
		})
	}

	return res
}

func ReadReplicaInputFromModel(r []ReadReplica) []client.ReadReplicaInput {
	var res []client.ReadReplicaInput
	for _, item := range r {
		res = append(res, client.ReadReplicaInput{
			Name: item.Name.ValueString(),
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

func ModelFromClient(postgres *client.Postgres, connectionInfo *client.PostgresConnectionInfo, logStreamOverrides *logs.ResourceLogStreamSetting, existingModel PostgresModel, diags diag.Diagnostics) PostgresModel {
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
		ReadReplicas:            ReadReplicaFromClient(postgres.ReadReplicas),
		Version:                 types.StringValue(string(postgres.Version)),
		ConnectionInfo:          connectionInfoFromClient(connectionInfo, diags),
		LogStreamOverride:       common.LogStreamOverrideFromClient(logStreamOverrides, existingModel.LogStreamOverride, diags),
		DiskSizeGB:              common.IntPointerAsValue(postgres.DiskSizeGB),
	}
	return postgresModel
}
