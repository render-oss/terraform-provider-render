package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/redis"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/redis/resource/internal"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &redisResource{}
	_ resource.ResourceWithConfigure        = &redisResource{}
	_ resource.ResourceWithImportState      = &redisResource{}
	_ resource.ResourceWithConfigValidators = &redisResource{}
)

// NewRedisResource is a helper function to simplify the provider implementation.
func NewRedisResource() resource.Resource {
	return &redisResource{}
}

// redisResource is the resource implementation.
type redisResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *redisResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *redisResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis"
}

// Schema defines the schema for the resource.
func (r *redisResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *redisResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan redis.RedisModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.CreateRedisRequestFromModel(r.ownerID, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service", "Could not create service, unexpected error: "+err.Error(),
		)
		return
	}

	var model client.Redis
	err = common.Create(func() (*http.Response, error) {
		return r.client.CreateRedis(ctx, serviceDetails)
	}, &model)
	if err != nil {
		resp.Diagnostics.AddError("Error creating service", err.Error())
		return
	}

	redisModel := redis.ModelForRedisResult(&model, resp.Diagnostics)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, redisModel)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *redisResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state redis.RedisModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var clientRedis client.Redis
	err := common.Get(func() (*http.Response, error) {
		return r.client.GetRedis(ctx, state.Id.ValueString())
	}, &clientRedis)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(state.Id.ValueString(), &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service", "Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	redisModel := redis.ModelForRedisResult(&clientRedis, resp.Diagnostics)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, redisModel)
	resp.Diagnostics.Append(diags...)
}

func (r *redisResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan redis.RedisModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state redis.RedisModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.UpdateServiceRequestFromModel(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service", "Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	var redisResponse client.Redis
	err = common.Update(func() (*http.Response, error) {
		return r.client.UpdateRedis(ctx, plan.Id.ValueString(), serviceDetails)
	}, &redisResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating redis", err.Error(),
		)
		return
	}

	redisModel := redis.ModelForRedisResult(&redisResponse, resp.Diagnostics)

	envID, err := common.UpdateEnvironmentID(ctx, r.client, redisModel.Id.ValueString(), &common.EnvironmentIDStateAndPlan{
		State: state.EnvironmentID.ValueStringPointer(),
		Plan:  plan.EnvironmentID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating environment ID", err.Error(),
		)
		return
	}

	redisModel.EnvironmentID = types.StringPointerValue(envID)

	diags = resp.State.Set(ctx, redisModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *redisResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state redis.RedisModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteRedis(ctx, state.Id.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting redis",
			err.Error(),
		)
		return
	}
}

func (r *redisResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *redisResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}
