package resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var NotificationOverride = schema.SingleNestedAttribute{
	Attributes: map[string]schema.Attribute{
		"notifications_to_send": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Description:         "The types of notifications to send. Must be one of default, all, failure, or none.",
			MarkdownDescription: "The types of notifications to send. Must be one of `default`, `all`, `failure`, or `none`.",
			Validators: []validator.String{
				stringvalidator.OneOf("default", "all", "failure", "none"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"preview_notifications_enabled": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Description:         "The types of notifications sent for previews of this service. Must be one of all, failure, or none.",
			MarkdownDescription: "Whether notifications for previews of this service are sent. Must be one of `all`, `failure`, or `none`.",
			Validators: []validator.String{
				stringvalidator.OneOf("default", "true", "false"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
	Optional:    true,
	Computed:    true,
	Description: "Configure the [notification settings](https://render.com/docs/notifications) for this service. These will override the global notification settings of the user or team.",
	PlanModifiers: []planmodifier.Object{
		objectplanmodifier.UseStateForUnknown(),
	},
}
