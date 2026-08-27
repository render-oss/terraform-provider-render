package redis

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
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

func ValidatePersistenceModeFunc() validator.String {
	return stringvalidator.OneOf(
		string(client.PersistenceModeJournalSnapshot),
		string(client.PersistenceModeSnapshot),
		string(client.PersistenceModeOff),
	)
}

func ValidateRedisPlanFunc() validator.String {
	return stringvalidator.Any(
		isNonCustomRedisPlanFunc(),
		isCustomRedisPlanFunc(),
	)
}

// isNonCustomRedisPlanFunc validates against the generated Redis plan values.
//
// A custom plan is named by its own identifier, matched by
// isCustomRedisPlanFunc, so the literal "custom" enum member is excluded.
func isNonCustomRedisPlanFunc() validator.String {
	return stringvalidator.OneOf(
		common.EnumStringsExcept(client.RedisPlanValues(), client.RedisPlanCustom)...,
	)
}

var customRegexp = regexp.MustCompile("^Custom.*$")

func isCustomRedisPlanFunc() validator.String {
	return stringvalidator.RegexMatches(customRegexp, "")
}
