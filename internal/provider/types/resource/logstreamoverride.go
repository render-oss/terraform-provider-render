package resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var LogStreamOverride = schema.SingleNestedAttribute{
	Attributes: map[string]schema.Attribute{
		"setting": schema.StringAttribute{
			Required:            true,
			Description:         "Whether to send or drop logs for this service. Must be one of `send` or `drop`.",
			MarkdownDescription: "Whether to send or drop logs for this service. Must be one of `send` or `drop`.",
			Validators: []validator.String{
				stringvalidator.OneOf("send", "drop"),
			},
		},
		"endpoint": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
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
	Optional:    true,
	Computed:    true,
	Description: "Configure the [log stream override settings](https://docs.render.com/log-streams#overriding-defaults) for this service. These will override the global log stream settings of the user or team.",
}
