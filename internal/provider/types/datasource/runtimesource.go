package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/common/validators"
)

type DockerDetailsModel struct {
	Context              types.String `tfsdk:"context"`
	DockerfilePath       types.String `tfsdk:"dockerfile_path"`
	RegistryCredentialID types.String `tfsdk:"registry_credential_id"`
}

var AutoDeploy = schema.BoolAttribute{
	Computed:            true,
	Default:             booldefault.StaticBool(true),
	Description:         "Automatic deploy on every push to your repository, or changes to your service settings or environment.",
	MarkdownDescription: "[Automatic deploy](https://docs.render.com/deploys#automatic-git-deploys) on every push to your repository, or changes to your service settings or environment.",
}

var DockerDetails = schema.SingleNestedAttribute{
	Computed:    true,
	Description: "Details for building and deploying a Dockerfile.",
	Attributes: map[string]schema.Attribute{
		"auto_deploy":  AutoDeploy,
		"repo_url":     RepoURL,
		"branch":       Branch,
		"build_filter": BuildFilter,
		"context": schema.StringAttribute{
			Computed:            true,
			Description:         "Docker build context directory. This is relative to your repository root. Defaults to the root.",
			MarkdownDescription: "[Docker build context directory.](https://docs.docker.com/reference/dockerfile/#usage) This is relative to your repository root. Defaults to the root.",
			Default:             stringdefault.StaticString("."),
		},
		"dockerfile_path": schema.StringAttribute{
			Computed:    true,
			Description: "Path to your Dockerfile relative to the repository root. This is not relative to your Docker build context. Example: `./subdir/Dockerfile.`",
			Default:     stringdefault.StaticString("./Dockerfile"),
		},
		"registry_credential_id": RegistryCredentialID,
	},
}

var NativeRuntimeDetails = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"auto_deploy":   AutoDeploy,
		"branch":        Branch,
		"build_command": BuildCommand,
		"build_filter":  BuildFilter,
		"repo_url":      RepoURL,
		"runtime":       Runtime,
	},
}

var ImageURL = schema.StringAttribute{
	CustomType:  commontypes.ImageURLStringType{},
	Computed:    true,
	Description: "URL of the Docker image to deploy.",
	Validators:  []validator.String{validators.StringNotEmpty},
}

var Image = schema.SingleNestedAttribute{
	Description: "Information about the pre-built image to deploy from a Docker registry.",
	Attributes: map[string]schema.Attribute{
		"image_url":              ImageURL,
		"registry_credential_id": RegistryCredentialID,
	},
}

var RegistryCredentialID = schema.StringAttribute{
	Description: "ID of the registry credential to use when pulling the image.",
	Computed:    true,
}

var RuntimeSource = schema.SingleNestedAttribute{
	Computed:    true,
	Description: "Source of the build artifacts or image that run your service.",
	Attributes: map[string]schema.Attribute{
		"native_runtime": NativeRuntimeDetails,
		"docker":         DockerDetails,
		"image":          RuntimeSourceImage,
	},
}

var PreDeployCommand = schema.StringAttribute{
	Description: "This command runs before starting your service. It is typically used for tasks like running a database migration or uploading assets to a CDN.",
	Computed:    true,
}

var ImageTag = schema.StringAttribute{
	Description: "Tag of the Docker image to deploy. Mutually exclusive with digest.",
	Optional:    true,
	Computed:    true,
}

var ImageDigest = schema.StringAttribute{
	Description: "Digest of the Docker image to deploy. Mutually exclusive with tag.",
	Optional:    true,
	Computed:    true,
}

var RuntimeSourceImage = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"image_url":              ImageURL,
		"tag":                    ImageTag,
		"digest":                 ImageDigest,
		"registry_credential_id": RegistryCredentialID,
	},
}
