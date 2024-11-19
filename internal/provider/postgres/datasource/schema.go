package datasource

import (
	"context"

	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/types/datasource"
	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func PostgresDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				Description:         "Unique identifier for this postgres",
				MarkdownDescription: "Unique identifier for this postgres",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         "Descriptive name for this postgres",
				MarkdownDescription: "Descriptive name for this postgres",
			},
			"datadog_api_key": schema.StringAttribute{
				Description:         "Datadog API key to use when sending postgres metrics",
				MarkdownDescription: "Datadog API key to use when sending postgres metrics",
				Optional:            true,
				Sensitive:           true,
			},
			"environment_id": datasource.EnvironmentID,
			"ip_allow_list":  datasource.IPAllowList,
			"database_name": schema.StringAttribute{
				CustomType:          commontypes.SuffixStringType{},
				Description:         "Name of the database in the postgres instance",
				MarkdownDescription: "Name of the database in the postgres instance",
				Computed:            true,
			},
			"database_user": schema.StringAttribute{
				Description:         "Name of the user in the postgres instance",
				MarkdownDescription: "Name of the user in the postgres instance",
				Computed:            true,
			},
			"high_availability_enabled": schema.BoolAttribute{
				Description:         "Whether high availability is enabled for this postgres",
				MarkdownDescription: "Whether high availability is enabled for this postgres",
				Computed:            true,
			},
			"plan": schema.StringAttribute{
				Description:         "Plan to use for this postgres",
				MarkdownDescription: "Plan to use for this postgres",
				Computed:            true,
			},
			"primary_postgres_id": schema.StringAttribute{
				Description:         "If this is a replica, the ID of the primary postgres instance",
				MarkdownDescription: "If this is a replica, the ID of the primary postgres instance",
				Computed:            true,
				Optional:            true,
			},
			"region": schema.StringAttribute{
				Description:         "Region the postgres instance in",
				MarkdownDescription: "Region the postgres instance in",
				Computed:            true,
			},
			"read_replicas": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name of the read replica.",
							MarkdownDescription: "Name of the read replica.",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							Description:         "ID of the read replica.",
							MarkdownDescription: "ID of the read replica.",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				Description:         "List of read replicas.",
				MarkdownDescription: "List of read replicas.",
			},
			"role": schema.StringAttribute{
				Description:         "Whether this postgres is a primary or replica",
				MarkdownDescription: "Whether this postgres is a primary or replica",
				Computed:            true,
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
				Computed:            true,
			},
			"log_stream_override": resource.LogStreamOverride,
			"disk_size_gb":        datasource.DiskSizeGB,
		},
	}
}
