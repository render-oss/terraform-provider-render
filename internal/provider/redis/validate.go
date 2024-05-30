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
		string(client.Free),
		string(client.Pro),
		string(client.ProPlus),
		string(client.Standard),
		string(client.Starter),
	)
}

var customRegexp = regexp.MustCompile("^Custom.*$")

func isCustomRedisPlanFunc() validator.String {
	return stringvalidator.RegexMatches(customRegexp, "")
}
