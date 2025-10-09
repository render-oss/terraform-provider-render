package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// IPAllowList is for datastores (postgres, redis, keyvalue) - always sends value to API
var IPAllowList = schema.SetNestedAttribute{
	NestedObject: IPAllowListItem,
	Optional:     true,
	Computed:     true,
	Description:  "List of IP addresses that are allowed to connect to the instance. If no IP addresses are provided, only connections via the private network will be allowed.",
}

// IPAllowListOptional is for webservices - state-aware management
// Omitted = don't manage (API default 0.0.0.0/0), Empty = block all, Values = allow those IPs
// Removing after being set = revert to default (0.0.0.0/0)
var IPAllowListOptional = schema.SetNestedAttribute{
	NestedObject: IPAllowListItem,
	Optional:     true,
	Description:  "List of IP addresses that are allowed to connect to the web service. If omitted, the API default (0.0.0.0/0 - allow all) is used. If set to an empty list, all traffic is blocked. If removed after being set, it reverts to the default (0.0.0.0/0). This is an enterprise-only feature.",
}
