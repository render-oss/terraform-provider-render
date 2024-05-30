package datasource

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

var Headers = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "Request paths to apply the header",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the header",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Description: "Value of the header",
			},
		},
	},
}
