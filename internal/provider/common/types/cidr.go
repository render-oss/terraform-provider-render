package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.ObjectTypable = CIDRType{}

type CIDRType struct {
	basetypes.ObjectType
}

func (t CIDRType) Equal(o attr.Type) bool {
	other, ok := o.(CIDRType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t CIDRType) String() string {
	return "CIDRType"
}

func (t CIDRType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	// CIDRValue defined in the value type section
	value := CIDRValue{
		ObjectValue: in,
	}

	return value, nil
}

func (t CIDRType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ObjectType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	objectValue, ok := attrValue.(basetypes.ObjectValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromObject(ctx, objectValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t CIDRType) ValueType(ctx context.Context) attr.Value {
	// CIDRValue defined in the value type section
	return CIDRValue{}
}

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.ObjectValuable = CIDRValue{}

type CIDRValue struct {
	basetypes.ObjectValue

	CidrBlock   string `tfsdk:"cidr_block"`
	Description string `tfsdk:"description"`
}

func (v CIDRValue) Equal(o attr.Value) bool {
	other, ok := o.(CIDRValue)

	if !ok {
		return false
	}

	return v.ObjectValue.Equal(other.ObjectValue)
}

func (v CIDRValue) Type(ctx context.Context) attr.Type {
	return CIDRType{}
}

func (v CIDRValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return v.ObjectValue, diag.Diagnostics{}
}
