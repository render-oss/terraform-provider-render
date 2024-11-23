package resource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-render/internal/provider/common/validators"
)

var Headers = schema.SetNestedAttribute{
	Optional:            true,
	Description:         "List of headers to apply to requests for static sites",
	MarkdownDescription: "List of [headers](https://render.com/docs/static-site-headers) to apply to requests for static sites",
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Request paths to apply the header",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the header",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "Value of the header",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
		},
	},
}
