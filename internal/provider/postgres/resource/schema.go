package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/common/validators"
	"terraform-provider-render/internal/provider/types/resource"
)

func PostgresResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Unique identifier for this postgres",
				MarkdownDescription: "Unique identifier for this postgres",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Descriptive name for this postgres",
				MarkdownDescription: "Descriptive name for this postgres",
				Validators:          []validator.String{validators.StringNotEmpty},
			},
			"datadog_api_key": schema.StringAttribute{
				Description:         "Datadog API key to use when sending postgres metrics",
				MarkdownDescription: "Datadog API key to use when sending postgres metrics",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Sensitive: true,
			},
			"environment_id": resource.ResourceEnvironmentID,
			"ip_allow_list":  resource.IPAllowList,
			"database_name": schema.StringAttribute{
				CustomType:          commontypes.SuffixStringType{},
				Description:         "Name of the database in the postgres instance",
				MarkdownDescription: "Name of the database in the postgres instance",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"database_user": schema.StringAttribute{
				Description:         "Name of the user in the postgres instance",
				MarkdownDescription: "Name of the user in the postgres instance",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"high_availability_enabled": schema.BoolAttribute{
				Description:         "Whether high availability is enabled for this postgres",
				MarkdownDescription: "Whether high availability is enabled for this postgres",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"plan": schema.StringAttribute{
				Description:         "Plan to use for this postgres. Must be one of free, starter, standard, pro, pro_plus, or a custom plan",
				MarkdownDescription: "Plan to use for this postgres. Must be one of `free`, `starter`, `standard`, `pro`, `pro_plus`, or a custom plan",
				Required:            true,
				Validators: []validator.String{
					ValidatePostgresPlanFunc(),
				},
			},
			"primary_postgres_id": schema.StringAttribute{
				Description:         "If this is a replica, the ID of the primary postgres instance",
				MarkdownDescription: "If this is a replica, the ID of the primary postgres instance",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Description:         "Region the postgres instance in",
				MarkdownDescription: "Region the postgres instance in",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{resource.RegionValidator},
			},
			"read_replicas": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name of the read replica.",
							MarkdownDescription: "Name of the read replica.",
							Required:            true,
						},
						"id": schema.StringAttribute{
							Description:         "ID of the read replica.",
							MarkdownDescription: "ID of the read replica.",
							Computed:            true,
						},
					},
				},
				Optional:            true,
				Description:         "List of read replicas.",
				MarkdownDescription: "List of read replicas.",
			},
			"role": schema.StringAttribute{
				Description:         "Whether this postgres is a primary or replica",
				MarkdownDescription: "Whether this postgres is a primary or replica",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_info": schema.SingleNestedAttribute{
				Description:         "Database connection info.",
				MarkdownDescription: "Database connection info.",
				Computed:            true,
				Sensitive:           true,
				Attributes: map[string]schema.Attribute{
					"password": schema.StringAttribute{
						Description:         "Password for the postgres user.",
						MarkdownDescription: "Password for the postgres user.",
						Computed:            true,
						Sensitive:           true,
					},
					"external_connection_string": schema.StringAttribute{
						Description:         "Connection string for external access. Use this to connect to the database from outside of Render.",
						MarkdownDescription: "Connection string for external access. Use this to connect to the database from outside of Render.",
						Computed:            true,
						Sensitive:           true,
					},
					"internal_connection_string": schema.StringAttribute{
						Description:         "Connection string for internal access. Use this to connect to the database from within the same Render region.",
						MarkdownDescription: "Connection string for internal access. Use this to connect to the database from within the same Render region.",
						Computed:            true,
						Sensitive:           true,
					},
					"psql_command": schema.StringAttribute{
						Description:         "Command to connect to the database using the `psql` command line tool.",
						MarkdownDescription: "Command to connect to the database using the `psql` command line tool.",
						Computed:            true,
						Sensitive:           true,
					},
				},
			},
			"version": schema.StringAttribute{
				Description:         "The Postgres version",
				MarkdownDescription: "The Postgres version",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					ValidatePostgresVersion(),
				},
			},
			"log_stream_override": resource.LogStreamOverride,
		},
	}
}
