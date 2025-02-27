package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client/eventtypes"
	"terraform-provider-render/internal/client/webhooks"
	"terraform-provider-render/internal/provider/webhook"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &webhookResource{}
	_ resource.ResourceWithConfigure = &webhookResource{}
)

// NewWebhookResource is a helper function to simplify the provider implementation.
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

// webhookResource is the resource implementation.
type webhookResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

// Schema defines the schema for the resource.
func (r *webhookResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhook.WebhookModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var whk webhooks.Webhook

	err := common.Create(func() (*http.Response, error) {
		var eventFilter webhooks.EventFilter
		for _, ef := range plan.EventFilter {
			eventFilter = append(eventFilter, eventtypes.EventType(ef.ValueString()))
		}

		return r.client.CreateWebhook(ctx, client.CreateWebhookJSONRequestBody{
			OwnerId: r.ownerID,

			Enabled:     plan.Enabled.ValueBool(),
			EventFilter: eventFilter,
			Name:        plan.Name.ValueString(),
			Url:         plan.URL.ValueString(),
		})
	}, &whk)
	if err != nil {
		resp.Diagnostics.AddError("unable to create webhook", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, webhook.WebhookFromClient(&whk, diags))
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	id, ok := common.IDFromState(ctx, req.State, resp.Diagnostics)
	if !ok {
		return
	}

	var state webhook.WebhookModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var whk webhooks.Webhook

	err := common.Get(func() (*http.Response, error) {
		return r.client.RetrieveWebhook(ctx, id)
	}, &whk)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(id, &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("unable to get webhook", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, webhook.WebhookFromClient(&whk, diags))
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhook.WebhookModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var whk webhooks.Webhook

	whkUpdate := r.webhookUpdate(plan)

	err := common.Update(func() (*http.Response, error) {
		return r.client.UpdateWebhook(ctx, plan.Id.ValueString(), whkUpdate)
	}, &whk)
	if err != nil {
		resp.Diagnostics.AddError("unable to create webhook", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, webhook.WebhookFromClient(&whk, diags))
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	id, ok := common.IDFromState(ctx, req.State, resp.Diagnostics)
	if !ok {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteWebhook(ctx, id)
	})
	if err != nil {
		resp.Diagnostics.AddError("unable to delete webhook", err.Error())
		return
	}
}

func (r *webhookResource) webhookUpdate(plan webhook.WebhookModel) client.UpdateWebhookJSONRequestBody {
	var eventFilter webhooks.EventFilter
	for _, ef := range plan.EventFilter {
		eventFilter = append(eventFilter, eventtypes.EventType(ef.ValueString()))
	}

	return client.UpdateWebhookJSONRequestBody{
		Enabled:     plan.Enabled.ValueBoolPointer(),
		EventFilter: &eventFilter,
		Name:        plan.Name.ValueStringPointer(),
		Url:         plan.URL.ValueStringPointer(),
	}
}
