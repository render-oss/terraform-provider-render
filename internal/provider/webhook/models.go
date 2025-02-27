package webhook

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client/webhooks"
)

type WebhookModel struct {
	Id          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	URL         types.String   `tfsdk:"url"`
	Enabled     types.Bool     `tfsdk:"enabled"`
	Secret      types.String   `tfsdk:"secret"`
	EventFilter []types.String `tfsdk:"event_filter"`
}

func WebhookFromClient(whk *webhooks.Webhook, diags diag.Diagnostics) *WebhookModel {
	var eventFilter []types.String
	for _, ef := range whk.EventFilter {
		eventFilter = append(eventFilter, types.StringValue(string(ef)))
	}

	return &WebhookModel{
		Id:          types.StringValue(whk.Id),
		Name:        types.StringValue(whk.Name),
		URL:         types.StringValue(whk.Url),
		Enabled:     types.BoolValue(whk.Enabled),
		Secret:      types.StringValue(whk.Secret),
		EventFilter: eventFilter,
	}
}
