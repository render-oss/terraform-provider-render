package datasource

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

var CustomDomain = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required:    true,
			Description: "Unique identifier for the custom domain",
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "DNS record of the custom domain",
		},
		"domain_type": schema.StringAttribute{
			Computed:    true,
			Description: "Type of the custom domain. Either apex or subdomain",
		},
		"public_suffix": schema.StringAttribute{
			Computed:    true,
			Description: "Public suffix of the custom domain",
		},
		"redirect_for_name": schema.StringAttribute{
			Computed:    true,
			Description: "DNS record of the custom domain to redirect to",
		},
	},
}

var CustomDomains = schema.SetNestedAttribute{
	Optional:     true,
	Description:  "Custom domains to associate with the service.",
	NestedObject: CustomDomain,
}
