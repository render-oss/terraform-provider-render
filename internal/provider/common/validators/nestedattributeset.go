package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = nestedAttributeSet{}

// nestedAttributeSet given a set of paths, checks that at least one of the paths has a non-null value.
type nestedAttributeSet struct {
	paths []path.Expression
}

func (v nestedAttributeSet) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v nestedAttributeSet) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that at least one attribute from this collection is set: %s", v.paths)
}

func (v nestedAttributeSet) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	expressions := request.PathExpression.MergeExpressions(v.paths...)

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
				continue
			}

			if mpVal.IsNull() {
				continue
			}

			return // At least one of the paths has a non-null value
		}
	}

	pathStrings := make([]string, len(v.paths))
	for i, p := range v.paths {
		pathStrings[i] = p.String()
	}

	response.Diagnostics.Append(validatordiag.InvalidBlockDiagnostic(
		request.Path,
		fmt.Sprintf("must have at least one of the following attributes set: [%s]", strings.Join(pathStrings, ", ")),
	))
}

// NestedAttributeSet checks that at least one nested attribute in a set of paths has a non-null value.
func NestedAttributeSet(paths ...path.Expression) validator.Object {
	return nestedAttributeSet{
		paths: paths,
	}
}
