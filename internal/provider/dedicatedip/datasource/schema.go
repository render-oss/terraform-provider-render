package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Schema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Dedicated IP.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique identifier for this dedicated IP.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Descriptive name for this dedicated IP.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Free-form description for this dedicated IP.",
			},
			"owner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the workspace that owns this dedicated IP.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "Region the dedicated IP applies in.",
			},
			"environment_ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Environments this dedicated IP applies to. Empty when the IP is workspace-scoped.",
			},
			"ips": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The IPv4 addresses assigned to this dedicated IP. Empty until provisioning completes (status is RUNNING).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Provisioning status. One of UNKNOWN, CREATING, PENDING, RUNNING, FAILED, DELETING, DELETED.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time the dedicated IP was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time the dedicated IP was last updated.",
			},
		},
	}
}
