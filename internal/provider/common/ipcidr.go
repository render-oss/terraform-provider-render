package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
)

const (
	CIDRBlockAttr   = "cidr_block"
	DescriptionAttr = "description"
)

var ipAllowListType = map[string]attr.Type{
	CIDRBlockAttr:   types.StringType,
	DescriptionAttr: types.StringType,
}

// AllowAllCIDRList is the default IP allow list that allows all traffic
var AllowAllCIDRList = []client.CidrBlockAndDescription{
	{
		CidrBlock:   "0.0.0.0/0",
		Description: "everywhere",
	},
}

func IPAllowListFromClient(c []client.CidrBlockAndDescription, diags diag.Diagnostics) types.Set {
	objType := types.ObjectType{AttrTypes: ipAllowListType}

	var res []attr.Value
	for _, item := range c {
		value, objectDiags := types.ObjectValue(
			ipAllowListType,
			map[string]attr.Value{
				CIDRBlockAttr:   types.StringValue(item.CidrBlock),
				DescriptionAttr: types.StringValue(item.Description),
			},
		)

		diags.Append(objectDiags...)
		if diags.HasError() {
			return types.SetNull(objType)
		}

		res = append(res, value)
	}

	setValue, setDiags := types.SetValue(objType, res)
	diags.Append(setDiags...)
	if diags.HasError() {
		return types.SetNull(objType)
	}

	return setValue
}

func ClientFromIPAllowList(c types.Set) ([]client.CidrBlockAndDescription, error) {
	// intentionally not using nil slice declaration to ensure that the struct is included
	// in PATCH requests and updates allowed_ips to nil when empty.
	res := []client.CidrBlockAndDescription{}
	for _, item := range c.Elements() {
		obj, ok := item.(types.Object)
		if !ok {
			return nil, fmt.Errorf("expected object type, got %T", item)
		}

		attrs := obj.Attributes()
		cidr, ok := attrs[CIDRBlockAttr]
		if !ok {
			return nil, fmt.Errorf("missing required attribute %s", CIDRBlockAttr)
		}
		cidrString, ok := cidr.(types.String)
		if !ok {
			return nil, fmt.Errorf("expected string type, got %T", cidr)
		}

		description, ok := attrs[DescriptionAttr]
		if !ok {
			return nil, fmt.Errorf("missing required attribute %s", description)
		}
		descriptionString, ok := description.(types.String)
		if !ok {
			return nil, fmt.Errorf("expected string type, got %T", description)
		}

		res = append(res, client.CidrBlockAndDescription{
			CidrBlock:   cidrString.ValueString(),
			Description: descriptionString.ValueString(),
		})
	}

	return res, nil
}
