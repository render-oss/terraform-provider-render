package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var ResourceEnvironmentID = schema.StringAttribute{
	Computed:    true,
	Description: "Unique identifier for the environment that the resource belongs to",
}

var EnvironmentID = schema.StringAttribute{
	Computed:    true,
	Description: "Unique identifier for the environment",
}

var EnvironmentName = schema.StringAttribute{
	Computed:    true,
	Description: "Name of the environment",
}

var EnvironmentProtectedStatus = schema.StringAttribute{
	Computed:    true,
	Description: "Protected environment",
}

var Environment = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"id":               EnvironmentID,
		"name":             EnvironmentName,
		"protected_status": EnvironmentProtectedStatus,
	},
}

var Environments = schema.MapNestedAttribute{
	Computed:     true,
	NestedObject: Environment,
	Description:  "Mapped list of environments",
	Validators: []validator.Map{
		mapvalidator.SizeAtLeast(1),
	},
}
