package validators

import "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

var StringNotEmpty = stringvalidator.LengthAtLeast(1)
