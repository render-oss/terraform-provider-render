package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/redis"
)

var KeyValuePlan = schema.StringAttribute{
	Optional:            true,
	Computed:            true,
	Description:         "Plan for the Key Value instance. Must be one of free, starter, standard, pro, pro_plus, or a custom plan.",
	MarkdownDescription: "Plan for the Key Value instance. Must be one of `free`, `starter`, `standard`, `pro`, `pro_plus`, or a custom plan.",
	Default:             stringdefault.StaticString(string(client.KeyValuePlanProPlus)),
	Validators: []validator.String{
		redis.ValidateRedisPlanFunc(),
	},
}
