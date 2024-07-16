package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/envgroup"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &envGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &envGroupDataSource{}
)

// NewEnvGroupDataSource is a helper function to simplify the provider implementation.
func NewEnvGroupDataSource() datasource.DataSource {
	return &envGroupDataSource{}
}

// envGroupDataSource is the data source implementation.common.
type envGroupDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *envGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *envGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_group"
}

// Schema defines the schema for the data source.
func (d *envGroupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = EnvGroupDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *envGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan envgroup.EnvGroupModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var envGroup client.EnvGroup
	err := common.Get(func() (*http.Response, error) {
		return d.client.RetrieveEnvGroup(ctx, plan.Id.ValueString())
	}, &envGroup)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get envGroup", err.Error())
		return
	}

	resp.State.Set(ctx, envgroup.ModelFromClient(&envGroup, nil))
}
