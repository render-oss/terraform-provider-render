package common

import (
	"terraform-provider-render/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CustomDomainModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	DomainType      types.String `tfsdk:"domain_type"`
	PublicSuffix    types.String `tfsdk:"public_suffix"`
	RedirectForName types.String `tfsdk:"redirect_for_name"`
}

func CustomDomainModelsToClientCustomDomains(customDomains []CustomDomainModel) []client.CustomDomain {
	clientCustomDomains := make([]client.CustomDomain, len(customDomains))
	for i, cd := range customDomains {
		clientCustomDomains[i] = client.CustomDomain{
			Id:              cd.Id.ValueString(),
			Name:            cd.Name.ValueString(),
			DomainType:      customDomainStringToClientType(cd.DomainType.ValueString()),
			PublicSuffix:    cd.PublicSuffix.ValueString(),
			RedirectForName: cd.RedirectForName.ValueString(),
		}
	}
	return clientCustomDomains
}

func CustomDomainClientsToCustomDomainModels(customDomains *[]client.CustomDomain) []CustomDomainModel {
	if customDomains == nil || len(*customDomains) == 0 {
		return nil
	}

	customDomainModels := make([]CustomDomainModel, len(*customDomains))
	for i, cd := range *customDomains {
		customDomainModels[i] = CustomDomainModel{
			Id:              types.StringValue(cd.Id),
			Name:            types.StringValue(cd.Name),
			DomainType:      types.StringValue(string(cd.DomainType)),
			PublicSuffix:    types.StringValue(cd.PublicSuffix),
			RedirectForName: types.StringValue(cd.RedirectForName),
		}
	}
	return customDomainModels
}

// CustomDomainClientsToCustomDomainModelsNonRedirecting returns only custom domains that are not auto-generated redirect domains.
// A domain is considered auto-generated if RedirectForName is non-empty.
func CustomDomainClientsToCustomDomainModelsNonRedirecting(customDomains *[]client.CustomDomain) []CustomDomainModel {
	if customDomains == nil || len(*customDomains) == 0 {
		return nil
	}
	filtered := make([]CustomDomainModel, 0, len(*customDomains))
	for _, cd := range *customDomains {
		if cd.RedirectForName != "" {
			continue
		}
		filtered = append(filtered, CustomDomainModel{
			Id:              types.StringValue(cd.Id),
			Name:            types.StringValue(cd.Name),
			DomainType:      types.StringValue(string(cd.DomainType)),
			PublicSuffix:    types.StringValue(cd.PublicSuffix),
			RedirectForName: types.StringValue(cd.RedirectForName),
		})
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func customDomainStringToClientType(domainType string) client.CustomDomainDomainType {
	switch domainType {
	case "apex":
		return client.CustomDomainDomainTypeApex
	case "subdomain":
		return client.CustomDomainDomainTypeSubdomain
	}
	return ""
}

var customDomainAttrTypes = map[string]attr.Type{
	"id":                types.StringType,
	"name":              types.StringType,
	"domain_type":       types.StringType,
	"public_suffix":     types.StringType,
	"redirect_for_name": types.StringType,
}

// CustomDomainSetFromClient converts client custom domains to a Terraform Set of Objects
// matching the CustomDomain schema. Returns a null set when input is nil or empty.
func CustomDomainSetFromClient(customDomains *[]client.CustomDomain, diags diag.Diagnostics) types.Set {
	objType := types.ObjectType{AttrTypes: customDomainAttrTypes}
	if customDomains == nil || len(*customDomains) == 0 {
		return types.SetNull(objType)
	}

	var elems []attr.Value
	for _, cd := range *customDomains {
		obj, oDiags := types.ObjectValue(customDomainAttrTypes, map[string]attr.Value{
			"id":                types.StringValue(cd.Id),
			"name":              types.StringValue(cd.Name),
			"domain_type":       types.StringValue(string(cd.DomainType)),
			"public_suffix":     types.StringValue(cd.PublicSuffix),
			"redirect_for_name": types.StringValue(cd.RedirectForName),
		})
		diags.Append(oDiags...)
		if diags.HasError() {
			return types.SetNull(objType)
		}
		elems = append(elems, obj)
	}

	set, sDiags := types.SetValue(objType, elems)
	diags.Append(sDiags...)
	if diags.HasError() {
		return types.SetNull(objType)
	}
	return set
}
