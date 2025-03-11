package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/provider/common/validators"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the webhook",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the webhook",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "whether or not the webhook is enabled",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "the URL to send webhooks to",
			},
			"event_filter": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Filter webhooks to only these events. If empty, all webhooks will be sent.",
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret to verify webhook signatures.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Description: "Provides a Render Webhook resource",
	}
}
