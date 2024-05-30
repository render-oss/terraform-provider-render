package resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var EnvVar = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Optional:  true,
			Computed:  true,
			Sensitive: true,
		},
		"generate_value": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "If true, Render will generate the variable value.",
			Default:     booldefault.StaticBool(false),
			Validators: []validator.Bool{
				boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value")),
				boolvalidator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("value")),
			},
		},
	},
}

var EnvVars = schema.MapNestedAttribute{
	NestedObject: EnvVar,
	Optional:     true,
	Description:  "Map of environment variable names to their values.",
	Validators: []validator.Map{
		mapvalidator.SizeAtLeast(1),
	},
}
