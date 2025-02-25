package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"metrics_provider": schema.StringAttribute{
				Required:    true,
				Description: "The metrics provider to use.",
				Validators: []validator.String{
					stringvalidator.OneOf("BETTER_STACK", "DATADOG", "GRAFANA", "HONEYCOMB", "NEW_RELIC", "CUSTOM"),
				},
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL to send metrics to.",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The token to use when sending metrics.",
			},
		},
		Description: "Configure the metrics stream settings for this owner.",
	}
}
