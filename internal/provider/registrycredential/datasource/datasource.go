package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &registryDataSource{}
	_ datasource.DataSourceWithConfigure = &registryDataSource{}
)

// NewRegistryDataSource is a helper function to simplify the provider implementation.
func NewRegistryDataSource() datasource.DataSource {
	return &registryDataSource{}
}

// registryDataSource is the data source implementation.common.
type registryDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *registryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *registryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_credential"
}

// Schema defines the schema for the data source.
func (d *registryDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RegistryCredentialDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *registryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan RegistryCredentialModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registryCredentials, err := d.client.RetrieveRegistryCredentialWithResponse(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get registry credentials", err.Error())
		return
	}

	if registryCredentials.StatusCode() != 200 {
		resp.Diagnostics.AddError("Unable to get registry credentials", registryCredentials.Status())
		return
	}

	creds := registryCredentials.JSON200
	registryCredentialsModel := RegistryCredentialModel{
		Id:       types.StringValue(creds.Id),
		Name:     types.StringValue(creds.Name),
		Registry: types.StringValue(string(creds.Registry)),
		Username: types.StringValue(creds.Username),
	}

	resp.State.Set(ctx, registryCredentialsModel)
}
