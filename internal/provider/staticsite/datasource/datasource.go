package datasource

import (
	"context"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/staticsite"
	rendertypes "terraform-provider-render/internal/provider/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &staticSiteSource{}
	_ datasource.DataSourceWithConfigure = &staticSiteSource{}
)

// NewStaticSiteSource is a helper function to simplify the provider implementation.
func NewStaticSiteSource() datasource.DataSource {
	return &staticSiteSource{}
}

// staticSiteSource is the data source implementation.common.
type staticSiteSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *staticSiteSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *staticSiteSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site"
}

// Schema defines the schema for the data source.
func (d *staticSiteSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *staticSiteSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan staticsite.StaticSiteModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	staticSite, err := d.client.GetServiceWithResponse(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get static site", err.Error())
		return
	}

	if staticSite.StatusCode() != 200 {
		resp.Diagnostics.AddError("Unable to get static site", staticSite.Status())
		return
	}

	wrappedService, err := common.WrapStaticSite(ctx, d.client, staticSite.JSON200)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get static site", err.Error())
		return
	}

	staticSitesModel, err := staticsite.ModelForServiceResult(wrappedService, plan, diags)
	if err != nil {
		resp.Diagnostics.AddError("Unable to apply service result to model", err.Error())
		return
	}

	resp.State.Set(ctx, staticSitesModel)
}
