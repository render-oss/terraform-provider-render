package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/provider/common/validators"
)

var ProjectID = schema.StringAttribute{
	Computed:    true,
	Description: "Unique identifier for the project",
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
	},
}

var ProjectName = schema.StringAttribute{
	Required:    true,
	Description: "Name of the project",
	Validators:  []validator.String{validators.StringNotEmpty},
}
