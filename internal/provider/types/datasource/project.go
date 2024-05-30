package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var ProjectID = schema.StringAttribute{
	Required:            true,
	Description:         "Unique identifier for the project",
	MarkdownDescription: "Unique identifier for the project",
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
	},
}

var ProjectName = schema.StringAttribute{
	Computed:            true,
	Description:         "Name of the project",
	MarkdownDescription: "Name of the project",
}
