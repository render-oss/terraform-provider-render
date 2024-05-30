package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var NotificationOverride = schema.SingleNestedAttribute{
	Attributes: map[string]schema.Attribute{
		"notifications_to_send": schema.StringAttribute{
			Computed:            true,
			Description:         "The types of notifications to send.",
			MarkdownDescription: "The types of notifications to send.",
			Validators: []validator.String{
				stringvalidator.OneOf("default", "all", "failure", "none"),
			},
		},
		"preview_notifications_enabled": schema.StringAttribute{
			Computed:            true,
			Description:         "Whether notifications for previews of this service are sent.",
			MarkdownDescription: "Whether notifications for previews of this service are sent.",
			Validators: []validator.String{
				stringvalidator.OneOf("default", "true", "false"),
			},
		},
	},
	Computed:    true,
	Description: "Set the notification settings for this service. These will override the notification settings of the owner.",
}
