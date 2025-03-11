package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique identifier of the webhook",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the webhook",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "whether or not the webhook is enabled",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "the URL to send webhooks to",
			},
			"event_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Filter webhooks to only these events. If empty, all webhooks will be sent.",
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret to verify webhook signatures.",
			},
		},
		Description: "Provides a Render Webhook datasource",
	}
}
