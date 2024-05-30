package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = otherKeyHasValue{}

// otherKeyHasValue validates that the value matches one of expected values.
type otherKeyHasValue struct {
	values   []string
	contains bool
	path     path.Expression
}

func (v otherKeyHasValue) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v otherKeyHasValue) MarkdownDescription(_ context.Context) string {
	if v.contains {
		return fmt.Sprintf("value at %q must be one of: %s", v.path.String(), strings.Join(v.values, ", "))
	}
	return fmt.Sprintf("value at %q must not be one of: %s", v.path.String(), strings.Join(v.values, ", "))
}

func (v otherKeyHasValue) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	expressions := request.PathExpression.MergeExpressions(v.path)

	for _, expression := range expressions {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)

		response.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, matchedPath := range matchedPaths {
			var mpVal attr.Value
			diags := request.Config.GetAttribute(ctx, matchedPath, &mpVal)
			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if mpVal.IsNull() {
				if v.contains {
					response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
						request.Path,
						v.Description(ctx),
						fmt.Sprintf("Attribute %q must be specified when %q is specified", matchedPath.String(), request.Path),
					))
				}
				return
			}

			for _, otherValue := range v.values {
				if mpVal.Equal(types.StringValue(otherValue)) {
					if v.contains {
						return
					}

					response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
						request.Path,
						v.Description(ctx),
						fmt.Sprintf("Attribute %q must not be one of: %s, when %q is specified", matchedPath.String(), strings.Join(v.values, ", "), request.Path),
					))

					return
				}
			}

			if v.contains {
				response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
					request.Path,
					v.Description(ctx),
					fmt.Sprintf("Attribute %q must be one of: %s, when %q is specified", matchedPath.String(), strings.Join(v.values, ", "), request.Path),
				))
				return
			}
		}
	}
}

// PathHasValue checks that the value at path. If contains is true
// then the value must be one of the values. If contains is false
// then the value must not be one of the values.
func PathHasValue(values []string, contains bool, path path.Expression) validator.Object {
	return otherKeyHasValue{
		values:   values,
		contains: contains,
		path:     path,
	}
}
