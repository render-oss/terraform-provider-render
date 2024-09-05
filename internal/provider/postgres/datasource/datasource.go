package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/provider/common"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/postgres"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &postgresDataSource{}
	_ datasource.DataSourceWithConfigure = &postgresDataSource{}
)

// NewPostgresDataSource is a helper function to simplify the provider implementation.
func NewPostgresDataSource() datasource.DataSource {
	return &postgresDataSource{}
}

// postgresDataSource is the data source implementation.common.
type postgresDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *postgresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *postgresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres"
}

// Schema defines the schema for the data source.
func (d *postgresDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PostgresDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *postgresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan postgres.PostgresModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pg client.Postgres
	err := common.Get(func() (*http.Response, error) {
		return d.client.RetrievePostgres(ctx, plan.ID.ValueString())
	}, &pg)
	if err != nil {
		resp.Diagnostics.AddError("unable to get postgres", err.Error())
		return
	}

	var secrets client.PostgresConnectionInfo
	if err = common.Get(func() (*http.Response, error) {
		return d.client.RetrievePostgresConnectionInfo(ctx, plan.ID.ValueString())
	}, &secrets); err != nil {
		resp.Diagnostics.AddError("unable to get postgres connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.GetLogStreamOverrides(ctx, d.client, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream overrides", err.Error())
		return
	}

	resp.State.Set(ctx, postgres.ModelFromClient(&pg, &secrets, logStreamOverrides, plan, resp.Diagnostics))
}
