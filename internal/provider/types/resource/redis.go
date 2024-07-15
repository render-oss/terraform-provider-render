package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/redis"
)

var MaxMemoryPolicy = schema.StringAttribute{
	Required:            true,
	Description:         "Policy for evicting keys when the maxmemory limit is reached. Valid values are allkeys_lfu, allkeys_lru, allkeys_random, noeviction, volatile_lfu, volatile_lru, volatile_random, volatile_ttl.",
	MarkdownDescription: "Policy for evicting keys when the maxmemory limit is reached. Valid values are `allkeys_lfu`, `allkeys_lru`, `allkeys_random`, `noeviction`, `volatile_lfu`, `volatile_lru`, `volatile_random`, `volatile_ttl.`",
	Validators: []validator.String{
		redis.ValidateMaxMemoryPolicyFunc(),
	},
}

var RedisPlan = schema.StringAttribute{
	Optional:            true,
	Computed:            true,
	Description:         "Plan for the Redis instance. Must be one of free, starter, standard, pro, pro_plus, or a custom plan.",
	MarkdownDescription: "Plan for the Redis instance. Must be one of `free`, `starter`, `standard`, `pro`, `pro_plus`, or a custom plan.",
	Default:             stringdefault.StaticString(string(client.RedisPlanProPlus)),
	Validators: []validator.String{
		redis.ValidateRedisPlanFunc(),
	},
}
