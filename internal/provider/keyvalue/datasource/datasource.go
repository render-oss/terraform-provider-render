package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/provider/common"

	"terraform-provider-render/internal/provider/keyvalue"

	"terraform-provider-render/internal/client"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &keyvalueSource{}
	_ datasource.DataSourceWithConfigure = &keyvalueSource{}
)

// NewKeyValueSource is a helper function to simplify the provider implementation.
func NewKeyValueSource() datasource.DataSource {
	return &keyvalueSource{}
}

// keyvalueSource is the data source implementation.common.
type keyvalueSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *keyvalueSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *keyvalueSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keyvalue"
}

// Schema defines the schema for the data source.
func (d *keyvalueSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *keyvalueSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan keyvalue.KeyValueModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var keyvalueResponse client.KeyValue
	if err := common.Get(func() (*http.Response, error) {
		return d.client.RetrieveKeyValue(ctx, plan.Id.ValueString())
	}, &keyvalueResponse); err != nil {
		resp.Diagnostics.AddError("unable to get keyvalue", err.Error())
		return
	}

	var connectionInfo client.KeyValueConnectionInfo
	if err := common.Get(func() (*http.Response, error) {
		return d.client.RetrieveKeyValueConnectionInfo(ctx, keyvalueResponse.Id)
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get keyvalue connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.GetLogStreamOverrides(ctx, d.client, keyvalueResponse.Id)
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream overrides", err.Error())
		return
	}

	keyvalueModel := keyvalue.ModelForKeyValueResult(&keyvalueResponse, &plan, &connectionInfo, logStreamOverrides, resp.Diagnostics)

	resp.State.Set(ctx, keyvalueModel)
}
