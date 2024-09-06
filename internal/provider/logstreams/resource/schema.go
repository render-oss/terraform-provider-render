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
			"preview": schema.StringAttribute{
				Required:            true,
				Description:         "Whether to send or drop logs for preview resources. Must be one of `send` or `drop`.",
				MarkdownDescription: "Whether to send or drop logs for preview resources. Must be one of `send` or `drop`.",
				Validators: []validator.String{
					stringvalidator.OneOf("send", "drop"),
				},
			},
			"endpoint": schema.StringAttribute{
				Optional:            true,
				Description:         "The endpoint to send logs to.",
				MarkdownDescription: "The endpoint to send logs to.",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "The token to use when sending logs.",
				MarkdownDescription: "The token to use when sending logs.",
			},
		},
		Description: "Configure the [log stream override settings](https://docs.render.com/log-streams#setup) for this owner.",
	}
}
