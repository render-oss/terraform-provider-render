package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/privateservice"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &privateServiceSource{}
	_ datasource.DataSourceWithConfigure = &privateServiceSource{}
)

// NewPrivateServiceSource is a helper function to simplify the provider implementation.
func NewPrivateServiceSource() datasource.DataSource {
	return &privateServiceSource{}
}

// privateServiceSource is the data source implementation.common.
type privateServiceSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *privateServiceSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *privateServiceSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_service"
}

// Schema defines the schema for the data source.
func (d *privateServiceSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *privateServiceSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan privateservice.PrivateServiceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := common.GetWrappedService(ctx, d.client, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get private service", err.Error())
		return
	}

	privateServiceModel, err := privateservice.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Unable to apply service result to model", err.Error())
		return
	}

	resp.State.Set(ctx, privateServiceModel)
}
