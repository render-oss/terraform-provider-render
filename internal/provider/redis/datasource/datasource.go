package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

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

	redisResponse, err := d.client.GetRedisWithResponse(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get service", err.Error())
		return
	}

	if redisResponse.StatusCode() != 200 {
		resp.Diagnostics.AddError("Unable to get service", err.Error())
		return
	}

	redisModel := redis.ModelForRedisResult(redisResponse.JSON200, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Unable to apply service result to model", err.Error())
		return
	}

	resp.State.Set(ctx, redisModel)
}
