package datasource

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

var PublishPath = schema.StringAttribute{
	Computed:    true,
	Description: "Path to the directory to publish",
}

var BuildCommand = schema.StringAttribute{
	Computed:    true,
	Description: "Command to build the service",
}

var StartCommand = schema.StringAttribute{
	Computed:    true,
	Description: "Command to run the service",
}

var Branch = schema.StringAttribute{
	Computed:    true,
	Description: "Branch to build",
}

var RepoURL = schema.StringAttribute{
	Computed:    true,
	Description: "URL of the repository to build",
}

var Routes = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				Computed:    true,
				Description: "Source path to match.",
			},
			"destination": schema.StringAttribute{
				Computed:    true,
				Description: "Destination path to route to.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of route. Either redirect or rewrite.",
			},
		},
	},
}
