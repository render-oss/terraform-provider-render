package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"metrics_provider": schema.StringAttribute{
				Computed:    true,
				Description: "The metrics provider to use.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL to send metrics to.",
			},
		},
		Description: "Configure the metrics stream settings for this owner.",
	}
}
