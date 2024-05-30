package resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var IPAllowListItem = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"cidr_block": schema.StringAttribute{
			Required:    true,
			Description: "CIDR block that is allowed to connect to the Redis instance. (0.0.0.0/0 to allow traffic from all IPs) ",
		},
		"description": schema.StringAttribute{
			Required:    true,
			Description: "Description of the IP address or range. This is used to help identify the IP address or range in the list.",
		},
	},
}

var IPAllowList = schema.SetNestedAttribute{
	NestedObject: IPAllowListItem,
	Optional:     true,
	Computed:     true,
	Description:  "List of IP addresses that are allowed to connect to the instance. If no IP addresses are provided, only connections via the private network will be allowed.",
	Validators: []validator.Set{
		setvalidator.SizeAtLeast(1),
	},
}
