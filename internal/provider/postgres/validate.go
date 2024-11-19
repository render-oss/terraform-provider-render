package postgres

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ValidateDiskSizeGB() DiskSizeGBValidator {
	return DiskSizeGBValidator{}
}

type DiskSizeGBValidator struct {
}

func (v DiskSizeGBValidator) Description(_ context.Context) string {
	return "value must be either 1 or a positive multiple of 5"
}
func (v DiskSizeGBValidator) MarkdownDescription(_ context.Context) string {
	return "value must be either 1 or a positive multiple of 5"
}
func (v DiskSizeGBValidator) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if request.ConfigValue.ValueInt64() == 1 {
		return
	}

	if request.ConfigValue.ValueInt64() <= 0 || request.ConfigValue.ValueInt64()%5 != 0 {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%d", request.ConfigValue.ValueInt64()),
		))
	}
}
