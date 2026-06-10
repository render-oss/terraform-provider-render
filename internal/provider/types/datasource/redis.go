package datasource

import (
	"terraform-provider-render/internal/provider/redis"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var MaxMemoryPolicy = schema.StringAttribute{
	Computed:    true,
	Description: "Policy for evicting keys when the maxmemory limit is reached",
	Validators: []validator.String{
		redis.ValidateMaxMemoryPolicyFunc(),
	},
}

var PersistenceMode = schema.StringAttribute{
	Computed:    true,
	Description: "The type of persistence to use for saving data",
	Validators: []validator.String{
		redis.ValidatePersistenceModeFunc(),
	},
}

var RedisPlan = schema.StringAttribute{
	Computed:    true,
	Description: "Plan for the Redis instance",
}
