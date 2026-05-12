package dedicatedip

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
)

// Model is the Terraform-side representation of a Dedicated IP.
// It is shared between the resource and the data source.
type Model struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	OwnerID        types.String `tfsdk:"owner_id"`
	Region         types.String `tfsdk:"region"`
	EnvironmentIDs types.Set    `tfsdk:"environment_ids"`
	IPs            types.List   `tfsdk:"ips"`
	Status         types.String `tfsdk:"status"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// ModelFromClient maps a wire-format DedicatedIP onto the Terraform model.
func ModelFromClient(d *client.DedicatedIP, diags *diag.Diagnostics) Model {
	envIDs, envDiags := stringSetFrom(d.EnvironmentIds)
	diags.Append(envDiags...)
	ips, ipDiags := stringListFrom(d.Ips)
	diags.Append(ipDiags...)

	m := Model{
		ID:             types.StringValue(d.Id),
		Name:           types.StringValue(d.Name),
		Description:    types.StringValue(d.Description),
		OwnerID:        types.StringValue(d.OwnerId),
		Region:         types.StringValue(string(d.Region)),
		EnvironmentIDs: envIDs,
		IPs:            ips,
		Status:         types.StringValue(string(d.Status)),
		CreatedAt:      types.StringValue(d.CreatedAt.Format(timeFormat)),
	}
	if d.UpdatedAt != nil {
		m.UpdatedAt = types.StringValue(d.UpdatedAt.Format(timeFormat))
	} else {
		m.UpdatedAt = types.StringNull()
	}
	return m
}

// timeFormat is RFC3339 — the same format the API emits.
const timeFormat = "2006-01-02T15:04:05Z07:00"

func stringListFrom(values []string) (types.List, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.ListValue(types.StringType, elems)
}

func stringSetFrom(values []string) (types.Set, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.SetValue(types.StringType, elems)
}

// EnvironmentIDsFromPlan extracts a non-nil []string from a Terraform set.
// With the schema's Default(empty set) the value should never be null or
// unknown in practice — the guard returns an empty slice defensively.
func EnvironmentIDsFromPlan(set types.Set, diags *diag.Diagnostics) []string {
	if set.IsNull() || set.IsUnknown() {
		return []string{}
	}
	out := make([]string, 0, len(set.Elements()))
	for _, e := range set.Elements() {
		s, ok := e.(types.String)
		if !ok {
			diags.AddError("Invalid environment_ids element", "expected string")
			return []string{}
		}
		out = append(out, s.ValueString())
	}
	return out
}
