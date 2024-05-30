package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var SecretFile = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"content": schema.StringAttribute{
			Computed:  true,
			Sensitive: true,
		},
	},
}

var SecretFiles = schema.MapNestedAttribute{
	NestedObject: SecretFile,
	Computed:     true,
	Description:  "A map of secret file paths to their contents.",
}
