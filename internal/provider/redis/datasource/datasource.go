package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"terraform-provider-render/internal/provider/common"

	"terraform-provider-render/internal/provider/redis"

	"terraform-provider-render/internal/client"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &redisSource{}
	_ datasource.DataSourceWithConfigure = &redisSource{}
)

// NewRedisSource is a helper function to simplify the provider implementation.
func NewRedisSource() datasource.DataSource {
	return &redisSource{}
}

// redisSource is the data source implementation.common.
type redisSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *redisSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *redisSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis"
}

// Schema defines the schema for the data source.
func (d *redisSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *redisSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan redis.RedisModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var redisResponse client.Redis
	if err := common.Get(func() (*http.Response, error) {
		return d.client.GetRedis(ctx, plan.Id.ValueString())
	}, &redisResponse); err != nil {
		resp.Diagnostics.AddError("unable to get redis", err.Error())
		return
	}

	var connectionInfo client.RedisConnectionInfo
	if err := common.Get(func() (*http.Response, error) {
		return d.client.GetRedisConnectionInfo(ctx, redisResponse.Id)
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get redis connection info", err.Error())
		return
	}

	redisModel := redis.ModelForRedisResult(&redisResponse, &connectionInfo, resp.Diagnostics)

	resp.State.Set(ctx, redisModel)
}
