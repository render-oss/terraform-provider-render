package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var Autoscaling = schema.SingleNestedAttribute{
	Attributes: map[string]schema.Attribute{
		"criteria": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"cpu": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"percentage": schema.Int64Attribute{
							Computed:    true,
							Description: "Determines when your service will be scaled. If the average resource utilization is significantly above/below the target, we will increase/decrease the number of instances.",
						},
					},
					Computed: true,
				},
				"memory": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"percentage": schema.Int64Attribute{
							Computed:    true,
							Description: "Determines when your service will be scaled. If the average resource utilization is significantly above/below the target, we will increase/decrease the number of instances.",
						},
					},
					Computed: true,
				},
			},
			Computed: true,
		},
		"enabled": schema.BoolAttribute{
			Computed: true,
		},
		"max": schema.Int64Attribute{
			Computed:    true,
			Description: "The maximum number of instances for the service",
		},
		"min": schema.Int64Attribute{
			Computed:    true,
			Description: "The minimum number of instances for the service",
		},
	},
	Computed: true,
}
