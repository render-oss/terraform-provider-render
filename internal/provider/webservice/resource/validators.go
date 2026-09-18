package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
)

// maintenanceModeFreeTierValidator rejects enabling maintenance mode on the free plan at
// plan time. The Render API only supports maintenance mode on non-free tiers (issue #80);
// without this, an explicit `maintenance_mode = { enabled = true }` on a free service is
// silently dropped from the request and surfaces later as a confusing
// "inconsistent result after apply" framework error instead of a clear message.
type maintenanceModeFreeTierValidator struct{}

func (v maintenanceModeFreeTierValidator) Description(_ context.Context) string {
	return "maintenance mode can only be enabled for non-free tier services"
}

func (v maintenanceModeFreeTierValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v maintenanceModeFreeTierValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var servicePlan types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("plan"), &servicePlan)...)
	if resp.Diagnostics.HasError() || servicePlan.IsNull() || servicePlan.IsUnknown() {
		return
	}
	if client.Plan(servicePlan.ValueString()) != client.PlanFree {
		return
	}

	enabledPath := path.Root("maintenance_mode").AtName("enabled")
	var enabled types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, enabledPath, &enabled)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if enabled.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			enabledPath,
			"Maintenance mode not available on the free plan",
			"maintenance mode can only be configured for non-free tier services; "+
				"remove maintenance_mode or upgrade the service plan.",
		)
	}
}
