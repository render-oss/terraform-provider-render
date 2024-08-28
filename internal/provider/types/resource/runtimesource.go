package resource

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/common/validators"
)

type DockerDetailsModel struct {
	Command              types.String `tfsdk:"command"`
	Context              types.String `tfsdk:"context"`
	DockerfilePath       types.String `tfsdk:"dockerfile_path"`
	RegistryCredentialID types.String `tfsdk:"registry_credential_id"`
}

var DockerDetails = schema.SingleNestedAttribute{
	Optional:            true,
	Description:         "Details for building and deploying a service using a Dockerfile.",
	MarkdownDescription: "Details for building and deploying a service [using a Dockerfile](https://docs.render.com/docker).",
	Attributes: map[string]schema.Attribute{
		"auto_deploy":  AutoDeploy,
		"repo_url":     RepoURL,
		"branch":       Branch,
		"build_filter": BuildFilter,
		"context": schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			Description:         "Docker build context directory. This is relative to your repository root. Defaults to the root.",
			MarkdownDescription: "[Docker build context directory.](https://docs.docker.com/reference/dockerfile/#usage) This is relative to your repository root. Defaults to the root.",
			Default:             stringdefault.StaticString("."),
		},
		"dockerfile_path": schema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Description: "Path to your Dockerfile relative to the repository root. This is not relative to your Docker build context. Example: `./subdir/Dockerfile.`",
			Default:     stringdefault.StaticString("./Dockerfile"),
		},
		"registry_credential_id": RegistryCredentialID,
	},
}

var NativeRuntimeDetails = schema.SingleNestedAttribute{
	Description:         "Details for building and deploying a service using one of Render's native runtimes.",
	MarkdownDescription: "Details for building and deploying a service using one of Render's [native runtimes](https://docs.render.com/native-runtimes).",
	Optional:            true,
	Attributes: map[string]schema.Attribute{
		"auto_deploy":   AutoDeploy,
		"branch":        Branch,
		"build_command": BuildCommand,
		"build_filter":  BuildFilter,
		"repo_url":      RepoURL,
		"runtime":       Runtime,
	},
	Validators: []validator.Object{objectvalidator.AlsoRequires(path.MatchRoot("start_command"))},
}

var ImageURL = schema.StringAttribute{
	CustomType:  commontypes.ImageURLStringType{},
	Required:    true,
	Description: "URL of the Docker image to deploy.",
	Validators: []validator.String{
		validators.StringNotEmpty,
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^(([^/:@]+)|(.+\/[^/:@]+))$`),
			"must not contain the tag or digest. Use the tag or digest fields instead",
		),
	},
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

var RegistryCredentialID = schema.StringAttribute{
	Description: "ID of the registry credential to use when pulling the image.",
	Optional:    true,
}

var RuntimeSource = schema.SingleNestedAttribute{
	Required:            true,
	Description:         "Source of the build artifacts or image that run your service. You must provide one of native_runtime, docker, or image.",
	MarkdownDescription: "Source of the build artifacts or image that run your service. You must provide one of [native_runtime](https://docs.render.com/native-runtimes), [docker](https://docs.render.com/docker), or [image](https://docs.render.com/deploy-an-image).",
	Attributes: map[string]schema.Attribute{
		"native_runtime": NativeRuntimeDetails,
		"docker":         DockerDetails,
		"image":          RuntimeSourceImage,
	},
}

var PreDeployCommand = schema.StringAttribute{
	Description: "This command runs before starting your service. It is typically used for tasks like running a database migration or uploading assets to a CDN.",
	Optional:    true,
}

var RuntimeSourceImage = schema.SingleNestedAttribute{
	Description:         "Details for deploying a service using a Docker image from a registry.",
	MarkdownDescription: "Details for deploying a service using a [Docker image from a registry](https://docs.render.com/deploy-an-image).",
	Optional:            true,
	Attributes: map[string]schema.Attribute{
		"image_url":              ImageURL,
		"tag":                    ImageTag,
		"digest":                 ImageDigest,
		"registry_credential_id": RegistryCredentialID,
	},
}

var RuntimeSourceValidator = resourcevalidator.ExactlyOneOf(
	path.MatchRoot("runtime_source").AtName("native_runtime"),
	path.MatchRoot("runtime_source").AtName("image"),
	path.MatchRoot("runtime_source").AtName("docker"),
)

var ImageTagOrDigestValidator = resourcevalidator.Conflicting(
	path.MatchRoot("runtime_source").AtName("image").AtName("tag"),
	path.MatchRoot("runtime_source").AtName("image").AtName("digest"),
)
