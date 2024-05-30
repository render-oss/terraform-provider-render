package resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var SecretFile = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"content": schema.StringAttribute{
			Description: "The content of the secret file.",
			Required:    true,
			Sensitive:   true,
		},
	},
}

var SecretFiles = schema.MapNestedAttribute{
	NestedObject: SecretFile,
	Optional:     true,
	Description:  "A map of secret file paths to their contents.",
	Validators: []validator.Map{
		mapvalidator.SizeAtLeast(1),
	},
}
