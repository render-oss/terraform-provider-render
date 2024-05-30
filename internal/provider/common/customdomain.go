package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
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

func customDomainStringToClientType(domainType string) client.CustomDomainDomainType {
	switch domainType {
	case "apex":
		return client.CustomDomainDomainTypeApex
	case "subdomain":
		return client.CustomDomainDomainTypeSubdomain
	}
	return ""
}
