package resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/provider/common/validators"
)

var ResourceEnvironmentID = schema.StringAttribute{
	Optional:            true,
	Description:         "ID of the project environment that the resource belongs to",
	MarkdownDescription: "ID of the [project environment](https://render.com/docs/projects) that the resource belongs to",
}

var EnvironmentID = schema.StringAttribute{
	Optional:    true,
	Computed:    true,
	Description: "Unique identifier of the environment",
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
	},
}

var EnvironmentName = schema.StringAttribute{
	Required:    true,
	Description: "Name of the environment",
	Validators:  []validator.String{validators.StringNotEmpty},
}

var EnvironmentProtectedStatus = schema.StringAttribute{
	Required:            true,
	Description:         "Protected environment status. One of protected, unprotected",
	MarkdownDescription: "Protected environment status. One of `protected`, `unprotected`",
	Validators: []validator.String{
		stringvalidator.OneOf(
			"protected",
			"unprotected",
		),
	},
}

var Environment = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"id":               EnvironmentID,
		"name":             EnvironmentName,
		"protected_status": EnvironmentProtectedStatus,
	},
}

var Environments = schema.MapNestedAttribute{
	Required:     true,
	NestedObject: Environment,
	Description:  "List of environments",
	Validators: []validator.Map{
		mapvalidator.SizeAtLeast(1),
	},
}
