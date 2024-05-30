package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"terraform-provider-render/internal/provider/redis"
)

var MaxMemoryPolicy = schema.StringAttribute{
	Computed:    true,
	Description: "Policy for evicting keys when the maxmemory limit is reached",
	Validators: []validator.String{
		redis.ValidateMaxMemoryPolicyFunc(),
	},
}

var RedisPlan = schema.StringAttribute{
	Computed:    true,
	Description: "Plan for the Redis instance",
}
