package redis

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/client"
)

func ValidateMaxMemoryPolicyFunc() validator.String {
	return stringvalidator.OneOf(
		string(client.AllkeysLfu),
		string(client.AllkeysLru),
		string(client.AllkeysRandom),
		string(client.Noeviction),
		string(client.VolatileLfu),
		string(client.VolatileLru),
		string(client.VolatileRandom),
		string(client.VolatileTtl),
	)
}

func ValidateRedisPlanFunc() validator.String {
	return stringvalidator.Any(
		isNonCustomRedisPlanFunc(),
		isCustomRedisPlanFunc(),
	)
}

func isNonCustomRedisPlanFunc() validator.String {
	return stringvalidator.OneOf(
		string(client.RedisPlanFree),
		string(client.RedisPlanPro),
		string(client.RedisPlanProPlus),
		string(client.RedisPlanStandard),
		string(client.RedisPlanStarter),
	)
}

var customRegexp = regexp.MustCompile("^Custom.*$")

func isCustomRedisPlanFunc() validator.String {
	return stringvalidator.RegexMatches(customRegexp, "")
}
