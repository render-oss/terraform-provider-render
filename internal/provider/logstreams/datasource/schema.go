package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"preview": schema.StringAttribute{
				Computed:            true,
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
		},
		Description: "Configure the [log stream override settings](https://render.com/docs/log-streams#setup) for this owner.",
	}
}
