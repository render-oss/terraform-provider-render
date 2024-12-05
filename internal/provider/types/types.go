package types

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
)

type Data struct {
	Client                       *client.ClientWithResponses
	OwnerID                      string
	Poller                       *common.Poller
	WaitForDeployCompletion      bool
	SkipDeployAfterServiceUpdate bool
}

func ConfigureDatasource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *Data {
	if req.ProviderData == nil {
		return nil
	}

	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return nil
	}

	return data
}

func ConfigureResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *Data {
	if req.ProviderData == nil {
		return nil
	}

	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return nil
	}

	return data
}
