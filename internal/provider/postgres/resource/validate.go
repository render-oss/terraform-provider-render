package resource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/client"
)

func ValidatePostgresVersion() validator.String {
	return stringvalidator.OneOf(
		string(client.N11),
		string(client.N12),
		string(client.N13),
		string(client.N14),
		string(client.N15),
		string(client.N16),
		string(client.N17),
	)
}
