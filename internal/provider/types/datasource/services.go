package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ServiceID = schema.StringAttribute{
	Required:            true,
	Description:         "Unique identifier for the service",
	MarkdownDescription: "Unique identifier for the service",
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
	},
}

var ServiceName = schema.StringAttribute{
	Computed:            true,
	Description:         "Name of the service",
	MarkdownDescription: "Name of the service",
}

var Slug = schema.StringAttribute{
	Computed:    true,
	Description: "Unique slug for the service",
}

var Runtime = schema.StringAttribute{
	Computed:    true,
	Description: "Runtime of the service to use.",
	Validators: []validator.String{
		stringvalidator.OneOf(
			"docker",
			"elixir",
			"go",
			"node",
			"python",
			"ruby",
			"rust",
			"image",
		),
	},
}

var Plan = schema.StringAttribute{
	Computed:    true,
	Description: "Plan to use for the service",
}

var Region = schema.StringAttribute{
	Computed:    true,
	Description: "Region to deploy the service",
	Validators: []validator.String{
		stringvalidator.OneOf(
			"frankfurt",
			"ohio",
			"oregon",
			"singapore",
			"virginia",
		),
	},
}

var HealthCheckPath = schema.StringAttribute{
	Computed:            true,
	Description:         "If you're running a server, enter the path where your server will always return a 200 OK response. We use it to monitor your app and for zero downtime deploys.",
	MarkdownDescription: "If you're running a server, enter the path where your server will always return a 200 OK response. We use it to monitor your app and for [zero downtime deploys](https://docs.render.com/deploys#zero-downtime-deploys).",
}

var NumInstances = schema.Int64Attribute{
	Computed: true,
}

var PRPreviewsEnabled = schema.BoolAttribute{
	Computed: true,
}

var ServiceURL = schema.StringAttribute{
	Computed:    true,
	Description: "URL that the service is accessible from.",
}

var Disk = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Unique identifier for the disk",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Name of the disk",
		},
		"size_gb": schema.Int64Attribute{
			Required:    true,
			Description: "Size of the disk in GB",
		},
		"mount_path": schema.StringAttribute{
			Required:    true,
			Description: "Absolute path to mount the disk.",
		},
	},
}

var BuildFilter = schema.SingleNestedAttribute{
	Computed:    true,
	Description: "Filter for files and paths to monitor for automatic deploys. Filter paths are absolute. If you've defined a root directory, you can still define paths outside of the root directory.",
	Attributes: map[string]schema.Attribute{
		"paths": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: "Changes that match these paths will trigger a new build.",
		},
		"ignored_paths": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: "Changes that match these paths will not trigger a new build.",
		},
	},
}

var CronJobSchedule = schema.StringAttribute{
	Computed:    true,
	Description: "Cron schedule to run the job",
}

var RootDirectory = schema.StringAttribute{
	Computed:    true,
	Description: "Defaults to repository root. When you specify a root directory that is different from your repository root, Render runs all your commands in the specified directory and ignores changes outside the directory.",
}
