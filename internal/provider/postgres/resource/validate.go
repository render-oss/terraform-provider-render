package resource

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/postgres"
)

func ValidatePostgresPlanFunc() validator.String {
	return stringvalidator.Any(
		isNonCustomPostgresPlanFunc(),
		isCustomPostgresPlanFunc(),
	)
}

func isNonCustomPostgresPlanFunc() validator.String {
	return stringvalidator.OneOf(
		string(postgres.Free),
		string(postgres.Starter),
		string(postgres.Standard),
		string(postgres.Pro),
		string(postgres.ProPlus),
		string(postgres.Custom),
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
