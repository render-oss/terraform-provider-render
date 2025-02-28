package metricstream

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client/metrics"
)

type MetricStreamSettingModel struct {
	Provider types.String `tfsdk:"metrics_provider"`
	URL      types.String `tfsdk:"url"`
	Token    types.String `tfsdk:"token"`
}

type DatasourceMetricStreamSettingModel struct {
	Provider types.String `tfsdk:"metrics_provider"`
	URL      types.String `tfsdk:"url"`
}

var metricStreamTypes = map[string]attr.Type{
	"metrics_provider": types.StringType,
	"url":              types.StringType,
	"token":            types.StringType,
}

var datasourceMetricStreamTypes = map[string]attr.Type{
	"metrics_provider": types.StringType,
	"url":              types.StringType,
}

func MetricStreamFromClient(client *metrics.MetricsStream, plan MetricStreamSettingModel, diags diag.Diagnostics) types.Object {
	if client == nil {
		return types.ObjectNull(metricStreamTypes)
	}

	objectValue, objectDiags := types.ObjectValue(
		metricStreamTypes,
		map[string]attr.Value{
			"metrics_provider": types.StringValue(string(client.Provider)),
			"url":              types.StringValue(client.Url),
			"token":            plan.Token,
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func DataSourceMetricStreamFromClient(client *metrics.MetricsStream, diags diag.Diagnostics) types.Object {
	if client == nil {
		return types.ObjectNull(datasourceMetricStreamTypes)
	}

	objectValue, objectDiags := types.ObjectValue(
		datasourceMetricStreamTypes,
		map[string]attr.Value{
			"metrics_provider": types.StringValue(string(client.Provider)),
			"url":              types.StringValue(client.Url),
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func ProviderFromPlan(provider string) metrics.OtelProviderType {
	switch provider {
	case "BETTER_STACK":
		return metrics.BETTERSTACK
	case "CUSTOM":
		return metrics.CUSTOM
	case "DATADOG":
		return metrics.DATADOG
	case "GRAFANA":
		return metrics.GRAFANA
	case "HONEYCOMB":
		return metrics.HONEYCOMB
	case "NEW_RELIC":
		return metrics.NEWRELIC
	default:
		return metrics.OtelProviderType(provider)
	}
}
