package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var EnvVar = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Computed:  true,
			Sensitive: true,
		},
		"generate_value": schema.BoolAttribute{
			Computed: true,
		},
	},
}

var EnvVars = schema.MapNestedAttribute{
	NestedObject: EnvVar,
	Computed:     true,
	Description:  "Map of environment variable names to their values.",
}
