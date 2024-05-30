package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
	"terraform-provider-render/internal/provider/webservice"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &webServiceSource{}
	_ datasource.DataSourceWithConfigure = &webServiceSource{}
)

// NewWebServiceSource is a helper function to simplify the provider implementation.
func NewWebServiceSource() datasource.DataSource {
	return &webServiceSource{}
}

// webServiceSource is the data source implementation.common.
type webServiceSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *webServiceSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *webServiceSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_service"
}

// Schema defines the schema for the data source.
func (d *webServiceSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *webServiceSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan webservice.WebServiceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := common.GetWrappedService(ctx, d.client, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get web service", err.Error())
		return
	}

	webServicesModel, err := webservice.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Unable to apply service result to model", err.Error())
		return
	}

	resp.State.Set(ctx, webServicesModel)
}
