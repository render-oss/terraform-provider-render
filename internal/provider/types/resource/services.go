package resource

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/provider/common/validators"
)

var ServiceID = schema.StringAttribute{
	Computed:    true,
	Description: "Unique identifier for the service",
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
	},
}

var Slug = schema.StringAttribute{
	Computed:    true,
	Description: "Unique slug for the service",
}

var ServiceName = schema.StringAttribute{
	Required:    true,
	Description: "Name of the service",
	Validators:  []validator.String{validators.StringNotEmpty},
}

var Runtime = schema.StringAttribute{
	Required:            true,
	Description:         "Runtime of the service to use. Must be one of elixir, go, node, python, ruby, rust.",
	MarkdownDescription: "Runtime of the service to use. Must be one of `elixir`, `go`, `node`, `python`, `ruby`, `rust`.",
	Validators: []validator.String{
		stringvalidator.OneOf(
			"elixir",
			"go",
			"node",
			"python",
			"ruby",
			"rust",
		),
	},
}

var CronJobSchedule = schema.StringAttribute{
	Required:    true,
	Description: "Cron schedule to run the job",
	Validators:  []validator.String{validators.StringNotEmpty},
}

var Plan = schema.StringAttribute{
	Required:            true,
	Description:         "Plan to use for the service. Must be one of starter, standard, pro, pro_plus, pro_max, pro_ultra, or a custom plan.",
	MarkdownDescription: "Plan to use for the service. Must be one of `starter`, `standard`, `pro`, `pro_plus`, `pro_max`, `pro_ultra`, or a custom plan.",
}

var RegionValidator = stringvalidator.OneOf(
	"frankfurt",
	"ohio",
	"oregon",
	"singapore",
	"virginia",
)

var ConnectionInfo = schema.SingleNestedAttribute{
	Description:         "Database connection info.",
	MarkdownDescription: "Database connection info.",
	Computed:            true,
	Sensitive:           true,
	Attributes: map[string]schema.Attribute{
		"external_connection_string": schema.StringAttribute{
			Description: "Connection string for external access. Use this to connect to the redis from outside of Render.",
			Computed:    true,
			Sensitive:   true,
		},
		"internal_connection_string": schema.StringAttribute{
			Description: "Connection string for internal access. Use this to connect to the redis from within the same Render region.",
			Computed:    true,
			Sensitive:   true,
		},
		"redis_cli_command": schema.StringAttribute{
			Description: "Command to connect to the redis using the redis command line tool.",
			Computed:    true,
			Sensitive:   true,
		},
	},
}

var Region = schema.StringAttribute{
	Required:            true,
	Description:         "Region to deploy the service. One of frankfurt, ohio, oregon, singapore, virginia.",
	MarkdownDescription: "[Region](https://docs.render.com/regions) to deploy the service. One of `frankfurt`, `ohio`, `oregon`, `singapore`, `virginia`.",
	Validators: []validator.String{
		RegionValidator,
	},
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	},
}

var HealthCheckPath = schema.StringAttribute{
	Optional:            true,
	Computed:            true,
	Description:         "If you're running a server, enter the path where your server will always return a 200 OK response. We use it to monitor your app and for zero downtime deploys.",
	MarkdownDescription: "If you're running a server, enter the path where your server will always return a 200 OK response. We use it to monitor your app and for [zero downtime deploys](https://docs.render.com/deploys#zero-downtime-deploys).",
	Default:             stringdefault.StaticString(""),
}

var NumInstances = schema.Int64Attribute{
	Optional:    true,
	Computed:    true,
	Description: "Number of replicas of the service to run. Defaults to 1 on service creation and current instance count on update. If you want to manage the service's instance count outside Terraform, leave num_instances unset.",
	Validators: []validator.Int64{
		int64validator.Between(1, 100),
	},
}

var PRPreviewsEnabled = schema.BoolAttribute{
	Optional:            true,
	Computed:            true,
	Default:             booldefault.StaticBool(false),
	Description:         "Enable pull request previews for the service.",
	MarkdownDescription: "Enable [pull request previews](https://docs.render.com/pull-request-previews#pull-request-previews-git-backed) for the service.",
}

var PublishPath = schema.StringAttribute{
	Optional:    true,
	Computed:    true,
	Default:     stringdefault.StaticString("public"),
	Description: "Path to the directory that contains the build artifacts to publish for a static site. Defaults to public/.",
}

var AutoDeploy = schema.BoolAttribute{
	Computed:            true,
	Optional:            true,
	Default:             booldefault.StaticBool(true),
	Description:         "Automatic deploy on every push to your repository, or changes to your service settings or environment.",
	MarkdownDescription: "[Automatic deploy](https://docs.render.com/deploys#automatic-git-deploys) on every push to your repository, or changes to your service settings or environment.",
}

var BuildCommand = schema.StringAttribute{
	Required:    true,
	Description: "Command to build the service",
}

var StartCommand = schema.StringAttribute{
	Optional:            true,
	Description:         "Command to run the service. When using native runtimes, this will be used as the start command. For Docker and image-backed services, this will override the default Docker command for the image.",
	MarkdownDescription: "Command to run the service. When using native runtimes, this will be used as the start command and is required. For [Docker](https://docs.render.com/docker) and [image-backed](https://docs.render.com/deploy-an-image) services, this will override the default Docker command for the image.",
}

var MaxShutdownDelaySeconds = schema.Int64Attribute{
	Optional:    true,
	Computed:    true,
	Description: "The maximum amount of time (in seconds) that Render waits for your application process to exit gracefully after sending it a SIGTERM signal before sending a SIGKILL signal.",
	Validators: []validator.Int64{
		int64validator.Between(1, 300),
	},
}

var Branch = schema.StringAttribute{
	Required:    true,
	Description: "Branch of the git repository to build.",
}

var RepoURL = schema.StringAttribute{
	Required:    true,
	Description: "URL of the git repository to build.",
	Validators:  []validator.String{validators.StringNotEmpty},
}

var RootDirectory = schema.StringAttribute{
	Computed:            true,
	Optional:            true,
	Description:         "When you specify a root directory, Render runs all your commands in the specified directory and ignores changes outside the directory. Defaults to the repository root.",
	MarkdownDescription: "When you specify a [root directory](https://docs.render.com/monorepo-support#root-directory), Render runs all your commands in the specified directory and ignores changes outside the directory. Defaults to the repository root.",
}

var Routes = schema.ListNestedAttribute{
	Optional:            true,
	Description:         "List of redirect and rewrite rules to apply to a static site.",
	MarkdownDescription: "List of [redirect and rewrite rules](https://docs.render.com/redirects-rewrites) to apply to a static site.",
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				Required:    true,
				Description: "Source path to match.",
			},
			"destination": schema.StringAttribute{
				Required:    true,
				Description: "Destination path to route to.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of route. Either redirect or rewrite.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"redirect",
						"rewrite",
					),
				},
			},
		},
	},
}

var ServiceURL = schema.StringAttribute{
	Computed:    true,
	Description: "URL that the service is accessible from.",
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
	},
}

var Disk = schema.SingleNestedAttribute{
	Description:         "Persistent disk to attach to the service.",
	MarkdownDescription: "[Persistent disk](https://docs.render.com/disks) to attach to the service.",
	Optional:            true,
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
			Validators: []validator.Int64{
				int64validator.Between(1, 1000),
			},
		},
		"mount_path": schema.StringAttribute{
			Required:    true,
			Description: "Absolute path to mount the disk.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(`^/.+`), "mount_path must be an absolute path starting with /"),
			},
		},
	},
}

type BuildFilterModel struct {
	IgnoredPaths []types.String `tfsdk:"ignored_paths"`
	Paths        []types.String `tfsdk:"paths"`
}

var BuildFilter = schema.SingleNestedAttribute{
	Optional:            true,
	Description:         "Apply build filters to configure which changes in your git repository trigger automatic deploys. If you've defined a root directory, you can still define paths outside of the root directory.",
	MarkdownDescription: "Apply [build filters](https://docs.render.com/monorepo-support#build-filters) to configure which changes in your git repository trigger automatic deploys. If you've defined a root directory, you can still define paths outside of the root directory.",
	Attributes: map[string]schema.Attribute{
		"paths": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Changes that match these paths will trigger a new build.",
		},
		"ignored_paths": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Changes that match these paths will not trigger a new build.",
		},
	},
	Validators: []validator.Object{
		validators.NestedAttributeSet(
			path.MatchRelative().AtName("paths"),
			path.MatchRelative().AtName("ignored_paths"),
		),
	},
}
