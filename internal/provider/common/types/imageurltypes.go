package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisfies the expected interfaces.
var _ basetypes.StringTypable = ImageURLStringType{}

type ImageURLStringType struct {
	basetypes.StringType
}

func (t ImageURLStringType) Equal(o attr.Type) bool {
	other, ok := o.(ImageURLStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t ImageURLStringType) String() string {
	return "ImageURLStringType"
}

func (t ImageURLStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := ImageURLStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t ImageURLStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t ImageURLStringType) ValueType(ctx context.Context) attr.Value {
	return ImageURLStringValue{}
}

type ImageURLStringValue struct {
	basetypes.StringValue
}

// Ensure the implementation satisfies the expected interfaces.
var _ basetypes.StringValuableWithSemanticEquals = ImageURLStringValue{}

func (v ImageURLStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(ImageURLStringValue)

	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	// The API may change the value of the attribute in a couple of ways:
	// 1. It may cut off the [http|https] `https://docker.io/library/nginx:latest` -> `docker.io/library/nginx:latest`
	// 2. It may expand it `nginx` -> `docker.io/library/nginx:latest`
	//
	// So we define semantic equality as a substring match in either direction.
	return strings.Contains(v.ValueString(), newValue.ValueString()) || strings.Contains(newValue.ValueString(), v.ValueString()), diags
}

func (v ImageURLStringValue) Equal(o attr.Value) bool {
	other, ok := o.(ImageURLStringValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v ImageURLStringValue) Type(ctx context.Context) attr.Type {
	return ImageURLStringType{}
}
