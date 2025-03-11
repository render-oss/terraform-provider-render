package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/client/webhooks"
	"terraform-provider-render/internal/provider/webhook"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &webhookDataSource{}
	_ datasource.DataSourceWithConfigure = &webhookDataSource{}
)

// NewWebhookDataSource is a helper function to simplify the provider implementation.
func NewWebhookDataSource() datasource.DataSource {
	return &webhookDataSource{}
}

// webhookDataSource is the data source implementation.common.
type webhookDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *webhookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *webhookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

// Schema defines the schema for the data source.
func (d *webhookDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *webhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan webhook.WebhookModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	var whk webhooks.Webhook

	err := common.Get(func() (*http.Response, error) {
		return d.client.RetrieveWebhook(ctx, plan.Id.ValueString())
	}, &whk)
	if err != nil {
		resp.Diagnostics.AddError("unable to get webhook", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, webhook.WebhookFromClient(&whk, diags))
	resp.Diagnostics.Append(diags...)
}
