package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"terraform-provider-render/internal/provider/logstreams"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &logStreamSettingDataSource{}
	_ datasource.DataSourceWithConfigure = &logStreamSettingDataSource{}
)

// NewLogStreamSettingDataSource is a helper function to simplify the provider implementation.
func NewLogStreamSettingDataSource() datasource.DataSource {
	return &logStreamSettingDataSource{}
}

// logStreamSettingDataSource is the data source implementation.common.
type logStreamSettingDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *logStreamSettingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *logStreamSettingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_stream"
}

// Schema defines the schema for the data source.
func (d *logStreamSettingDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *logStreamSettingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var settings logstreams.DatasourceLogStreamSettingModel
	diags := req.Config.Get(ctx, &settings)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	var logStreamSetting client.GetOwnerLogStreamResponse

	err := common.Get(func() (*http.Response, error) {
		return d.client.GetOwnerLogStream(ctx, d.ownerID)
	}, &logStreamSetting)
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream settings", err.Error())
		return
	}
	if logStreamSetting.JSON200 == nil {
		resp.Diagnostics.AddError("unable to get log stream settings", "not found")
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, logstreams.DataSourceLogStreamFromClient(logStreamSetting.JSON200, diags))
	resp.Diagnostics.Append(diags...)
}
