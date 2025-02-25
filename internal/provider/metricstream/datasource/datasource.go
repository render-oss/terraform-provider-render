package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"terraform-provider-render/internal/client/metrics"
	"terraform-provider-render/internal/provider/metricstream"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &metricStreamSettingDataSource{}
	_ datasource.DataSourceWithConfigure = &metricStreamSettingDataSource{}
)

// NewMetricsStreamSettingDataSource is a helper function to simplify the provider implementation.
func NewMetricsStreamSettingDataSource() datasource.DataSource {
	return &metricStreamSettingDataSource{}
}

// metricStreamSettingDataSource is the data source implementation.common.
type metricStreamSettingDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *metricStreamSettingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *metricStreamSettingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics_stream"
}

// Schema defines the schema for the data source.
func (d *metricStreamSettingDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *metricStreamSettingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var settings metricstream.DatasourceMetricStreamSettingModel
	diags := req.Config.Get(ctx, &settings)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	var metricStreamSetting metrics.MetricsStream

	err := common.Get(func() (*http.Response, error) {
		return d.client.GetOwnerMetricsStream(ctx, d.ownerID)
	}, &metricStreamSetting)
	if err != nil {
		resp.Diagnostics.AddError("unable to get metric stream settings", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, metricstream.DataSourceMetricStreamFromClient(&metricStreamSetting, diags))
	resp.Diagnostics.Append(diags...)
}
