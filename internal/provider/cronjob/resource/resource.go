package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	cronJob "terraform-provider-render/internal/provider/cronjob"
	"terraform-provider-render/internal/provider/cronjob/resource/internal"
	rendertypes "terraform-provider-render/internal/provider/types"
	resourcecommon "terraform-provider-render/internal/provider/types/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &cronJobResource{}
	_ resource.ResourceWithConfigure        = &cronJobResource{}
	_ resource.ResourceWithImportState      = &cronJobResource{}
	_ resource.ResourceWithConfigValidators = &cronJobResource{}
)

// NewCronJobResource is a helper function to simplify the provider implementation.
func NewCronJobResource() resource.Resource {
	return &cronJobResource{}
}

// cronJobResource is the resource implementation.
type cronJobResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *cronJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *cronJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cron_job"
}

// Schema defines the schema for the resource.
func (r *cronJobResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *cronJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cronJob.CronJobModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.CreateServiceRequestFromModel(r.ownerID, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service", "Could not create service, unexpected error: "+err.Error(),
		)
		return
	}

	service, err := common.CreateService(ctx, r.client, common.CreateServiceReq{
		Service:              serviceDetails,
		EnvironmentID:        plan.EnvironmentID.ValueStringPointer(),
		NotificationOverride: plan.NotificationOverride,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service", "Could not create service, unexpected error: "+err.Error())
		return
	}

	model, err := cronJob.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cron job", "Could not create cron job, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, *model)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *cronJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cronJob.CronJobModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := common.GetWrappedService(ctx, r.client, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service", "Could not read service, unexpected error: "+err.Error(),
		)
		return
	}
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(state.Id.ValueString(), &diags)
		resp.State.RemoveResource(ctx)
		return
	}

	cronJobModel, err := cronJob.ModelForServiceResult(service, state, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service", "Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, cronJobModel)
	resp.Diagnostics.Append(diags...)
}

func (r *cronJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cronJob.CronJobModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state cronJob.CronJobModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.UpdateServiceRequestFromModel(plan, r.ownerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service", "Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	evs, err := common.EnvVarsToClient(plan.EnvVars)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service", "Could not process env vars, unexpected error: "+err.Error(),
		)
		return
	}

	notificationOverride, err := common.NotificationOverrideToClient(plan.NotificationOverride)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service", "Could not process notification override, unexpected error: "+err.Error(),
		)
		return
	}

	service, err := common.UpdateService(ctx, r.client, common.UpdateServiceReq{
		ServiceID:   plan.Id.ValueString(),
		Service:     serviceDetails,
		EnvVars:     evs,
		SecretFiles: common.SecretFilesToClient(plan.SecretFiles),
		EnvironmentID: &common.EnvironmentIDStateAndPlan{
			State: state.EnvironmentID.ValueStringPointer(),
			Plan:  plan.EnvironmentID.ValueStringPointer(),
		},
		NotificationOverride: notificationOverride,
		LogStreamOverride: &common.LogStreamOverrideStateAndPlan{
			State: state.LogStreamOverride,
			Plan:  plan.LogStreamOverride,
		},
	}, common.ServiceTypeCronJob)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service",
			"Could not update service, unexpected error: "+err.Error(),
		)
		return
	}
	cronJobModel, err := cronJob.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service",
			"Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, cronJobModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cronJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cronJob.CronJobModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteService(ctx, state.Id.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting service",
			err.Error(),
		)
		return
	}
}

func (r *cronJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cronJobResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcecommon.RuntimeSourceValidator,
		resourcecommon.ImageTagOrDigestValidator,
	}
}
