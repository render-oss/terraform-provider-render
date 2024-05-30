package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
)

type HeaderModel struct {
	Path  types.String `tfsdk:"path"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func HeaderResponseToClientHeaders(headerResp []client.HeaderWithCursor) []client.Header {
	headers := make([]client.Header, len(headerResp))
	for i, hr := range headerResp {
		headers[i] = client.Header{
			Path:  hr.Header.Path,
			Name:  hr.Header.Name,
			Value: hr.Header.Value,
		}
	}
	return headers
}

func ModelToClientHeaderInput(headerModels []HeaderModel) []client.HeaderInput {
	headers := make([]client.HeaderInput, len(headerModels))
	for i, hr := range headerModels {
		headers[i] = client.HeaderInput{
			Path:  hr.Path.ValueString(),
			Name:  hr.Name.ValueString(),
			Value: hr.Value.ValueString(),
		}
	}
	return headers
}

func ClientHeadersToRouteModels(header *[]client.Header) []HeaderModel {
	if header == nil || len(*header) == 0 {
		return nil
	}
	headers := make([]HeaderModel, len(*header))
	for i, h := range *header {
		headers[i] = HeaderModel{
			Path:  types.StringValue(h.Path),
			Name:  types.StringValue(h.Name),
			Value: types.StringValue(h.Value),
		}
	}

	return headers
}
