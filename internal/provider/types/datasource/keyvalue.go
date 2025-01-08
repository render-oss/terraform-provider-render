package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var KeyValuePlan = schema.StringAttribute{
	Computed:    true,
	Description: "Plan for the Key Value instance",
}
