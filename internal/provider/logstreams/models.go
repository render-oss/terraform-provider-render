package logstreams

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client/logs"
)

type LogStreamSettingModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
	Preview  types.String `tfsdk:"preview"`
}

type DatasourceLogStreamSettingModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Preview  types.String `tfsdk:"preview"`
}

var logStreamTypes = map[string]attr.Type{
	"preview":  types.StringType,
	"token":    types.StringType,
	"endpoint": types.StringType,
}

var datasourceLogStreamTypes = map[string]attr.Type{
	"preview":  types.StringType,
	"endpoint": types.StringType,
}

func LogStreamFromClient(client *logs.OwnerLogStreamSetting, plan LogStreamSettingModel, diags diag.Diagnostics) types.Object {
	if client == nil {
		return types.ObjectNull(logStreamTypes)
	}

	preview := logs.LogStreamPreviewSettingSend
	if client != nil {
		preview = *client.Preview
	}

	endpoint := types.StringNull()
	if client.Endpoint != nil && *client.Endpoint != "" {
		endpoint = types.StringValue(*client.Endpoint)
	}

	objectValue, objectDiags := types.ObjectValue(
		logStreamTypes,
		map[string]attr.Value{
			"preview":  types.StringValue(string(preview)),
			"endpoint": endpoint,
			"token":    plan.Token,
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func DataSourceLogStreamFromClient(client *logs.OwnerLogStreamSetting, diags diag.Diagnostics) types.Object {
	if client == nil {
		return types.ObjectNull(datasourceLogStreamTypes)
	}

	preview := logs.LogStreamPreviewSettingSend
	if client != nil {
		preview = *client.Preview
	}

	endpoint := ""
	if client.Endpoint != nil {
		endpoint = *client.Endpoint
	}

	objectValue, objectDiags := types.ObjectValue(
		datasourceLogStreamTypes,
		map[string]attr.Value{
			"preview":  types.StringValue(string(preview)),
			"endpoint": types.StringValue(endpoint),
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}
