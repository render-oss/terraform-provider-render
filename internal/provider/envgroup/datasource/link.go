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
	_ datasource.DataSource              = &envGroupLinkDataSource{}
	_ datasource.DataSourceWithConfigure = &envGroupLinkDataSource{}
)

// NewEnvGroupLinkDataSource is a helper function to simplify the provider implementation.
func NewEnvGroupLinkDataSource() datasource.DataSource {
	return &envGroupLinkDataSource{}
}

// envGroupDataSource is the data source implementation.common.
type envGroupLinkDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *envGroupLinkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *envGroupLinkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_group_link"
}

// Schema defines the schema for the data source.
func (d *envGroupLinkDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = EnvGroupLinkDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *envGroupLinkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan envgroup.EnvGroupLinkModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var envGroupLink client.EnvGroup
	err := common.Get(func() (*http.Response, error) {
		return d.client.GetEnvGroup(ctx, plan.EnvGroupId.ValueString())
	}, &envGroupLink)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get environment variable group link", err.Error())
		return
	}

	resp.State.Set(ctx, envgroup.LinkModelFromClient(&envGroupLink))
}
