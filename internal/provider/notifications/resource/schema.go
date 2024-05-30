package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NotificationSettingResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether email notifications are enabled.",
			},
			"notifications_to_send": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The types of notifications to send. Must be one of all, failure, or none.",
				MarkdownDescription: "The types of notifications to send. Must be one of `all`, `failure`, or `none`.",
				Validators: []validator.String{
					stringvalidator.OneOf("all", "failure", "none"),
				},
			},
			"preview_notifications_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether notifications for preview environments are sent.",
			},
			"slack_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Slack notifications are enabled.",
			},
		},
	}
}
