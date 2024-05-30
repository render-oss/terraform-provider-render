package resource

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"terraform-provider-render/internal/client"
)

func ValidatePostgresPlanFunc() validator.String {
	return stringvalidator.Any(
		isNonCustomPostgresPlanFunc(),
		isCustomPostgresPlanFunc(),
	)
}

func isNonCustomPostgresPlanFunc() validator.String {
	return stringvalidator.OneOf(
		string(client.PostgresPlansFree),
		string(client.PostgresPlansStarter),
		string(client.PostgresPlansStandard),
		string(client.PostgresPlansPro),
		string(client.PostgresPlansProPlus),
		string(client.PostgresPlansCustom),
	)
}

var customRegexp = regexp.MustCompile("^Custom.*$")

func isCustomPostgresPlanFunc() validator.String {
	return stringvalidator.RegexMatches(customRegexp, "")
}

func ValidatePostgresVersion() validator.String {
	return stringvalidator.OneOf(
		string(client.N11),
		string(client.N12),
		string(client.N13),
		string(client.N14),
		string(client.N15),
		string(client.N16),
	)
}
