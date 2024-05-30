package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/cronjob"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cronJobSource{}
	_ datasource.DataSourceWithConfigure = &cronJobSource{}
)

// NewCronJobSource is a helper function to simplify the provider implementation.
func NewCronJobSource() datasource.DataSource {
	return &cronJobSource{}
}

// cronJobSource is the data source implementation.common.
type cronJobSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *cronJobSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *cronJobSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cron_job"
}

// Schema defines the schema for the data source.
func (d *cronJobSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *cronJobSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan cronJob.CronJobModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := common.GetWrappedService(ctx, d.client, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get cron job", err.Error())
		return
	}

	cronJobModel, err := cronJob.ModelForServiceResult(service, plan.EnvVars, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Unable to apply service result to model", err.Error())
		return
	}

	resp.State.Set(ctx, cronJobModel)
}
