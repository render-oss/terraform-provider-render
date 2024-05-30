package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var Autoscaling = schema.SingleNestedAttribute{
	Attributes: map[string]schema.Attribute{
		"criteria": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"cpu": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Required: true,
						},
						"percentage": schema.Int64Attribute{
							Required:    true,
							Description: "Determines when your service will be scaled. If the average resource utilization is significantly above/below the target, we will increase/decrease the number of instances.",
						},
					},
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Object{
						objectplanmodifier.UseStateForUnknown(),
					},
				},
				"memory": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Required: true,
						},
						"percentage": schema.Int64Attribute{
							Required:    true,
							Description: "Determines when your service will be scaled. If the average resource utilization is significantly above/below the target, we will increase/decrease the number of instances.",
						},
					},
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Object{
						objectplanmodifier.UseStateForUnknown(),
					},
				},
			},
			Required: true,
		},
		"enabled": schema.BoolAttribute{
			Required: true,
		},
		"max": schema.Int64Attribute{
			Required:            true,
			Description:         "The maximum number of instances for the service",
			MarkdownDescription: "The maximum number of instances for the service",
		},
		"min": schema.Int64Attribute{
			Required:            true,
			Description:         "The minimum number of instances for the service",
			MarkdownDescription: "The minimum number of instances for the service",
		},
	},
	Optional: true,
}
