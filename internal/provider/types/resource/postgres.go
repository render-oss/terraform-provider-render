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
	Description:         "Plan to use for this postgres. Must be free, a basic plan (like basic_256mb), a pro plan (like pro_4gb), an accelerated plan (like accelerated_16gb), a custom plan, or one of: " + common.EnumList(specPostgresPlans()),
	MarkdownDescription: "Plan to use for this postgres. Must be `free`, a basic plan (like `basic_256mb`), a pro plan (like `pro_4gb`), an accelerated plan (like `accelerated_16gb`), a custom plan, or one of: " + common.EnumListMarkdown(specPostgresPlans()),
	Required:            true,
	Validators: []validator.String{
		ValidatePostgresPlanFunc(),
	},
}

// specPlanRegexp matches the plan names that spell out their compute, {cpu}c-{ram},
// such as 1c-4g. The remaining names are grouped into families by the
// description above rather than listed value by value.
var specPlanRegexp = regexp.MustCompile(`^[\d.]+c-\d+(mb|g)$`)

func specPostgresPlans() []postgres.PostgresPlans {
	return common.EnumValuesMatching(postgres.PostgresPlansValues(), specPlanRegexp)
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
