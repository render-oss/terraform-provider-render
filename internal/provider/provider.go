package provider

import (
	"context"
	"net/http"
	"os"

	projectdatasource "terraform-provider-render/internal/provider/project/datasource"
	projectresource "terraform-provider-render/internal/provider/project/resource"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-render/internal/provider/common"

	backgroundwokrerresource "terraform-provider-render/internal/provider/backgroundworker/resource"
	envgroupresource "terraform-provider-render/internal/provider/envgroup/resource"
	notificationsdatasource "terraform-provider-render/internal/provider/notifications/datasource"
	privateserviceresource "terraform-provider-render/internal/provider/privateservice/resource"
	redisdatasource "terraform-provider-render/internal/provider/redis/datasource"
	redisresource "terraform-provider-render/internal/provider/redis/resource"
	webservicedatasource "terraform-provider-render/internal/provider/webservice/datasource"
	webserviceresource "terraform-provider-render/internal/provider/webservice/resource"

	"terraform-provider-render/internal/client"
	backgroundworkerdatasource "terraform-provider-render/internal/provider/backgroundworker/datasource"
	cronjobdatasource "terraform-provider-render/internal/provider/cronjob/datasource"
	cronjobresource "terraform-provider-render/internal/provider/cronjob/resource"
	envgroupdatasource "terraform-provider-render/internal/provider/envgroup/datasource"
	notificationsresource "terraform-provider-render/internal/provider/notifications/resource"
	postgresdatasource "terraform-provider-render/internal/provider/postgres/datasource"
	postgresresource "terraform-provider-render/internal/provider/postgres/resource"
	privateservicedatasource "terraform-provider-render/internal/provider/privateservice/datasource"
	registrycredentialdatasource "terraform-provider-render/internal/provider/registrycredential/datasource"
	registrycredentialresource "terraform-provider-render/internal/provider/registrycredential/resource"
	staticsitedatasource "terraform-provider-render/internal/provider/staticsite/datasource"
	staticsiteresource "terraform-provider-render/internal/provider/staticsite/resource"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// renderProviderModel maps provider schema data to a Go type.
type renderProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	OwnerID types.String `tfsdk:"owner_id"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider              = &renderProvider{}
	_ provider.ProviderWithFunctions = &renderProvider{}
)

type ConfigFunc func(provider *renderProvider)

func WithHost(host string) ConfigFunc {
	return func(p *renderProvider) {
		p.Host = host
	}
}

func WithAPIKey(key string) ConfigFunc {
	return func(p *renderProvider) {
		p.APIKey = key
	}
}

func WithOwnerID(ownerID string) ConfigFunc {
	return func(p *renderProvider) {
		p.OwnerID = ownerID
	}
}

func WithHTTPClient(client *http.Client) ConfigFunc {
	return func(p *renderProvider) {
		p.httpClient = client
	}
}

func WithPoller(poller *common.Poller) ConfigFunc {
	return func(p *renderProvider) {
		p.poller = poller
	}
}

func WithWaitForDeployCompletion(wait bool) ConfigFunc {
	return func(p *renderProvider) {
		p.waitForDeployCompletion = wait
	}
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string, configFuncs ...ConfigFunc) func() provider.Provider {
	return func() provider.Provider {
		p := &renderProvider{
			version:                 version,
			waitForDeployCompletion: true,
		}

		for _, f := range configFuncs {
			f(p)
		}

		return p
	}
}

// renderProvider is the provider implementation.
type renderProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version                 string
	APIKey                  string `tfsdk:"api_key"`
	OwnerID                 string `tfsdk:"owner_id"`
	Host                    string
	httpClient              *http.Client
	poller                  *common.Poller
	waitForDeployCompletion bool
}

// Metadata returns the provider type name.
func (p *renderProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "render"
	resp.Version = p.version
}

var renderProviderDescription = ` The Render provider is used to interact with and manage resources on Render. The provider requires an API key and owner ID to be used.

The provider is currently in beta and may have breaking changes in the future.`

// Schema defines the provider-level schema for configuration data.
func (p *renderProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: renderProviderDescription,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "API key to use when interacting with the API. You can generate an API key from the user settings on the Render dashboard. The provider will read this value from the RENDER_API_KEY environment variable if set. This key is sensitive and should not be committed to source control.",
			},
			"owner_id": schema.StringAttribute{
				Optional:    true,
				Description: "The user or team ID that owns the managed resources. All resources will be created under this owner ID. You can find the owner ID in the Render dashboard by navigating to the user or team settings and finding the ID in the URL. The ID will start with usr- for individual accounts and tea- for team accounts. The provider will read this value from the RENDER_OWNER_ID environment variable if set.",
			},
		},
	}
}

// Configure prepares a Render API Client for data sources and resources.
func (p *renderProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Render Client")

	// Retrieve provider data from configuration
	var config renderProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown API Key",
			"The provider cannot create the Render API client as there is an unknown configuration value for the Render API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the RENDER_API_KEY environment variable.",
		)
	}

	if config.OwnerID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Owner ID",
			"The provider cannot create the Render API Client as there is an unknown configuration value for the Render API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the RENDER_OWNER_ID environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	if p.poller == nil {
		p.poller = &common.DefaultPoller
	}

	if p.Host == "" {
		p.Host = os.Getenv("RENDER_HOST")
	}
	if p.APIKey == "" {
		p.APIKey = os.Getenv("RENDER_API_KEY")
	}
	if p.OwnerID == "" {
		p.OwnerID = os.Getenv("RENDER_OWNER_ID")
	}

	if !config.APIKey.IsNull() {
		p.APIKey = config.APIKey.ValueString()
	}

	if !config.OwnerID.IsNull() {
		p.OwnerID = config.OwnerID.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if p.APIKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Render API Key",
			"The provider cannot create the Render API Client as there is a missing or empty value for the Render API Key. "+
				"Set the host value in the configuration or use the RENDER_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if p.OwnerID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("owner_id"),
			"Missing Render Owner ID",
			"The provider cannot create the Render API Client as there is a missing or empty value for the owner ID. "+
				"Set the username value in the configuration or use the RENDER_OWNER_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if p.Host == "" {
		p.Host = "https://api.render.com/v1"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "render_api_key", p.APIKey)
	ctx = tflog.SetField(ctx, "render_owner_id", p.OwnerID)
	ctx = tflog.SetField(ctx, "render_host", p.Host)

	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "render_api_key")

	tflog.Debug(ctx, "Creating Render Client")

	opts := []client.ClientOption{
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+p.APIKey)
			req.Header.Set("User-Agent", "terraform-provider-render/"+p.version)
			return nil
		}),
	}

	if p.httpClient != nil {
		opts = append(opts, client.WithHTTPClient(p.httpClient))
	}

	client, err := client.NewClientWithResponses(
		p.Host,
		opts...,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Render API Client",
			"An unexpected error occurred when creating the Render API Client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Render Client Error: "+err.Error(),
		)

		tflog.Info(ctx, "Configured Render Client", map[string]any{"success": true})
		return

	}

	data := &rendertypes.Data{
		Client:                  client,
		OwnerID:                 p.OwnerID,
		Poller:                  p.poller,
		WaitForDeployCompletion: p.waitForDeployCompletion,
	}

	resp.DataSourceData = data
	resp.ResourceData = data
}

// DataSources defines the data sources implemented in the provider.
func (p *renderProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		backgroundworkerdatasource.NewBackgroundWorkerSource,
		cronjobdatasource.NewCronJobSource,
		envgroupdatasource.NewEnvGroupDataSource,
		envgroupdatasource.NewEnvGroupLinkDataSource,
		notificationsdatasource.NewNotificationSettingDataSource,
		postgresdatasource.NewPostgresDataSource,
		privateservicedatasource.NewPrivateServiceSource,
		projectdatasource.NewProjectDataSource,
		redisdatasource.NewRedisSource,
		registrycredentialdatasource.NewRegistryDataSource,
		staticsitedatasource.NewStaticSiteSource,
		webservicedatasource.NewWebServiceSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *renderProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		backgroundwokrerresource.NewBackgroundWorkerResource,
		cronjobresource.NewCronJobResource,
		envgroupresource.NewEnvGroupResource,
		envgroupresource.NewEnvGroupLinkResource,
		notificationsresource.NewNotificationSettingResource,
		postgresresource.NewPostgresResource,
		privateserviceresource.NewPrivateServiceResource,
		projectresource.NewProjectResource,
		redisresource.NewRedisResource,
		registrycredentialresource.NewRegistryCredentialResource,
		staticsiteresource.NewStaticSiteResource,
		webserviceresource.NewWebServiceResource,
	}
}

func (p *renderProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}
