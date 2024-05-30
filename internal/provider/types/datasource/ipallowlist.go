package datasource

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

var IPAllowListItem = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"cidr_block": schema.StringAttribute{
			Computed:    true,
			Description: "CIDR block that is allowed to connect to the Redis instance. (0.0.0.0/0 to allow traffic from all IPs) ",
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: "Description of the IP address or range. This is used to help identify the IP address or range in the list.",
		},
	},
}

var IPAllowList = schema.SetNestedAttribute{
	NestedObject: IPAllowListItem,
	Computed:     true,
	Description:  "List of IP addresses that are allowed to connect to the Redis instance. If no IP addresses are provided, only connections via the private network will be allowed.",
}
