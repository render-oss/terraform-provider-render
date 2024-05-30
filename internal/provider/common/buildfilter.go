package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
)

func ClientBuildFilterForModel(buildFilter *BuildFilterModel) *client.BuildFilter {
	if buildFilter == nil {
		// Return an empty filter if the model is nil to ensure a PATCH request sets the filter to nil.
		return &client.BuildFilter{}
	}

	var ignoredPaths []string
	for _, p := range buildFilter.IgnoredPaths {
		ignoredPaths = append(ignoredPaths, p.ValueString())
	}

	var paths []string
	for _, p := range buildFilter.Paths {
		paths = append(paths, p.ValueString())
	}

	return &client.BuildFilter{
		IgnoredPaths: ignoredPaths,
		Paths:        paths,
	}
}

func BuildFilterModelForClient(buildFilter *client.BuildFilter) *BuildFilterModel {
	if buildFilter == nil {
		return nil
	}

	var ignoredPaths []types.String
	for _, p := range buildFilter.IgnoredPaths {
		ignoredPaths = append(ignoredPaths, types.StringValue(p))
	}

	var paths []types.String
	for _, p := range buildFilter.Paths {
		paths = append(paths, types.StringValue(p))
	}

	return &BuildFilterModel{
		IgnoredPaths: ignoredPaths,
		Paths:        paths,
	}
}
