package resource

import (
	"regexp"

	"terraform-provider-render/internal/client/postgres"
	"terraform-provider-render/internal/provider/common"
	providerpostgres "terraform-provider-render/internal/provider/postgres"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var PostgresPlan = schema.StringAttribute{
	Description:         "Plan to use for this postgres. Must be free, a basic plan (like basic_256mb), a pro plan (like pro_4gb), an accelerated plan (like accelerated_16gb), starter, standard, pro, pro_plus, or a custom plan",
	MarkdownDescription: "Plan to use for this postgres. Must be `free`, a basic plan (like `basic_256mb`), a pro plan (like `pro_4gb`), an accelerated plan (like `accelerated_16gb`), `starter`, `standard`, `pro`, `pro_plus`, or a custom plan",
	Required:            true,
	Validators: []validator.String{
		ValidatePostgresPlanFunc(),
	},
}

func ValidatePostgresPlanFunc() validator.String {
	return stringvalidator.Any(
		isNonCustomPostgresPlanFunc(),
		isCustomPostgresPlanFunc(),
	)
}

func isNonCustomPostgresPlanFunc() validator.String {
	// A custom plan is named by its own identifier, matched by
	// isCustomPostgresPlanFunc, so the literal "custom" enum member is not a
	// plan a config can ask for.
	return stringvalidator.OneOf(
		common.EnumStringsExcept(postgres.PostgresPlansValues(), postgres.Custom)...,
	)
}

var customRegexp = regexp.MustCompile("^Custom.*$")

func isCustomPostgresPlanFunc() validator.String {
	return stringvalidator.RegexMatches(customRegexp, "")
}

var DiskSizeGB schema.Int64Attribute = schema.Int64Attribute{
	Description: "Disk size in GB.",
	Computed:    true,
	Optional:    true,
	Validators:  []validator.Int64{providerpostgres.ValidateDiskSizeGB()},
}
