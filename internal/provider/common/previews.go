package common

import (
	"context"

	"terraform-provider-render/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// PreviewsToPreviewsObject generates a types.Object because `previews` is Generated and Optional, so Terraform needs
// the types.Object so it knows if it is Unknown versus Null.
func PreviewsToPreviewsObject(previews *client.Previews) types.Object {
	attributeTypes := map[string]attr.Type{
		"generation": types.StringType,
	}

	attributes := map[string]attr.Value{}
	if previews == nil {
		attributes["generation"] = types.StringNull()
	} else {
		attributes["generation"] = types.StringPointerValue((*string)(previews.Generation))
	}

	return types.ObjectValueMust(attributeTypes, attributes)
}

func PreviewsObjectToPreviews(ctx context.Context, previews types.Object) *client.Previews {
	previewsModel := &PreviewsModel{}
	previews.As(ctx, previewsModel, basetypes.ObjectAsOptions{})

	p := &client.Previews{}
	if previewsModel.Generation.IsNull() || previewsModel.Generation.IsUnknown() {
		return p
	}
	generation := previewsModel.Generation.ValueString()
	p.Generation = (*client.PreviewsGeneration)(&generation)
	return p
}
