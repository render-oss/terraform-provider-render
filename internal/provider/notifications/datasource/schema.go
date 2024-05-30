package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NotificationSettingDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email_enabled": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether email notifications are enabled.",
				MarkdownDescription: "Whether email notifications are enabled.",
			},
			"notifications_to_send": schema.StringAttribute{
				Computed:            true,
				Description:         "The types of notifications to send.",
				MarkdownDescription: "The types of notifications to send.",
				Validators: []validator.String{
					stringvalidator.OneOf("all", "failure", "none"),
				},
			},
			"preview_notifications_enabled": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether notifications for preview environments are sent.",
				MarkdownDescription: "Whether notifications for preview environments are sent.",
			},
			"slack_enabled": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether Slack notifications are enabled.",
				MarkdownDescription: "Whether Slack notifications are enabled.",
			},
		},
	}
}
