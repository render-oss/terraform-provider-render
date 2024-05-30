package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func IDFromState(ctx context.Context, state tfsdk.State, diagnostics diag.Diagnostics) (string, bool) {
	var id string
	diags := state.GetAttribute(ctx, path.Root("id"), &id)
	diagnostics.Append(diags...)

	return id, !diagnostics.HasError()
}

func EnvVarsFromState(ctx context.Context, state tfsdk.State, diagnostics diag.Diagnostics) (map[string]EnvVarModel, bool) {
	var evs map[string]EnvVarModel
	diags := state.GetAttribute(ctx, path.Root("env_vars"), &evs)
	diagnostics.Append(diags...)

	return evs, !diagnostics.HasError()
}
