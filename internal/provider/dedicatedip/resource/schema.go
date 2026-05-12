package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	rendertypes "terraform-provider-render/internal/provider/types/resource"
)

func Schema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Dedicated IP resource. Use this to allocate a workspace-scoped or environment-scoped egress IP that services in the same region will route outbound traffic through.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for this dedicated IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Descriptive name for this dedicated IP.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Free-form description for this dedicated IP.",
				Default:     stringdefault.StaticString(""),
			},
			"owner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the workspace that owns this dedicated IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": rendertypes.Region,
			// environment_ids defaults to the empty set so omit / null / [] are
			// all equivalent — and so dashboard drift (someone adds env scoping
			// out-of-band) surfaces as a plan diff that apply overwrites.
			"environment_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Environments to scope this dedicated IP to. Leave unset (or pass an empty set) to apply the IP to every service in the workspace within its region. Mutually exclusive with another workspace-scoped IP in the same region.",
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
			"ips": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The IPv4 addresses assigned to this dedicated IP. Empty until provisioning completes (when status is RUNNING).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Provisioning status. One of UNKNOWN, CREATING, PENDING, RUNNING, FAILED, DELETING, DELETED.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time the dedicated IP was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time the dedicated IP was last updated.",
			},
		},
	}
}
