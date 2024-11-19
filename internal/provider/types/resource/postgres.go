package resource

import (
	"regexp"
	"terraform-provider-render/internal/client/postgres"
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
	return stringvalidator.OneOf(
		string(postgres.Free),

		string(postgres.Basic256mb),
		string(postgres.Basic1gb),
		string(postgres.Basic4gb),

		string(postgres.Pro4gb),
		string(postgres.Pro8gb),
		string(postgres.Pro16gb),
		string(postgres.Pro32gb),
		string(postgres.Pro64gb),
		string(postgres.Pro128gb),
		string(postgres.Pro192gb),
		string(postgres.Pro256gb),
		string(postgres.Pro384gb),
		string(postgres.Pro512gb),

		string(postgres.Accelerated16gb),
		string(postgres.Accelerated32gb),
		string(postgres.Accelerated64gb),
		string(postgres.Accelerated128gb),
		string(postgres.Accelerated256gb),
		string(postgres.Accelerated384gb),
		string(postgres.Accelerated512gb),
		string(postgres.Accelerated768gb),
		string(postgres.Accelerated1024gb),

		string(postgres.Starter),
		string(postgres.Standard),
		string(postgres.Pro),
		string(postgres.ProPlus),
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
