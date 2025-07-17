package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/backgroundworker"
	"terraform-provider-render/internal/provider/backgroundworker/resource/internal"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
	resourcecommon "terraform-provider-render/internal/provider/types/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &backgroundWorkerResource{}
	_ resource.ResourceWithConfigure        = &backgroundWorkerResource{}
	_ resource.ResourceWithImportState      = &backgroundWorkerResource{}
	_ resource.ResourceWithConfigValidators = &backgroundWorkerResource{}
)

// NewBackgroundWorkerResource is a helper function to simplify the provider implementation.
func NewBackgroundWorkerResource() resource.Resource {
	return &backgroundWorkerResource{}
}

// backgroundWorkerResource is the resource implementation.
type backgroundWorkerResource struct {
	client                       *client.ClientWithResponses
	ownerID                      string
	poller                       *common.Poller
	waitForDeployCompletion      bool
	skipDeployAfterServiceUpdate bool
}

// Configure adds the provider configured Client to the resource.
func (r *backgroundWorkerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
	r.poller = data.Poller
	r.waitForDeployCompletion = data.WaitForDeployCompletion
	r.skipDeployAfterServiceUpdate = data.SkipDeployAfterServiceUpdate
}

// Metadata returns the resource type name.
func (r *backgroundWorkerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_background_worker"
}

// Schema defines the schema for the resource.
func (r *backgroundWorkerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *backgroundWorkerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var shouldWaitForServiceCompletion = r.waitForDeployCompletion
	var plan backgroundWorker.BackgroundWorkerModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.CreateServiceRequestFromModel(ctx, r.ownerID, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service", "Could not create service, unexpected error: "+err.Error(),
		)
		return
	}

	_, err = common.CreateService(ctx, r.client, common.CreateServiceReq{
		Service:              serviceDetails,
		EnvironmentID:        plan.EnvironmentID.ValueStringPointer(),
		NotificationOverride: plan.NotificationOverride,
		LogStreamOverride:    plan.LogStreamOverride,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating background worker", "Could not create background worker, unexpected error: "+err.Error(),
		)
	}

	service, err := common.GetWrappedServiceByName(ctx, r.client, r.ownerID, plan.Name.ValueString(), client.BackgroundWorker)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating background worker", "Could not find service, unexpected error: "+err.Error(),
		)
		return
	}

	res, err := backgroundWorker.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating background worker",
			err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, *res)
	resp.Diagnostics.Append(diags...)

	if !shouldWaitForServiceCompletion {
		return
	}

	// Wait for the service to be ready before returning
	err = common.WaitForService(ctx, r.poller, r.client, service.Id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating background worker",
			"Service never started: "+err.Error(),
		)
		return
	}
}

// Read resource information.
func (r *backgroundWorkerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan backgroundWorker.BackgroundWorkerModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := common.GetWrappedService(ctx, r.client, plan.Id.ValueString())
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(plan.Id.ValueString(), &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service",
			"Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	backgroundWorkerModel, err := backgroundWorker.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service",
			"Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, backgroundWorkerModel)
	resp.Diagnostics.Append(diags...)
}

func (r *backgroundWorkerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backgroundWorker.BackgroundWorkerModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state backgroundWorker.BackgroundWorkerModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.UpdateServiceRequestFromModel(ctx, plan, r.ownerID)
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

	service, err := common.UpdateService(ctx, r.client, r.skipDeployAfterServiceUpdate, common.UpdateServiceReq{
		ServiceID:   plan.Id.ValueString(),
		Service:     serviceDetails,
		EnvVars:     evs,
		SecretFiles: common.SecretFilesToClient(plan.SecretFiles),
		Disk: &common.DiskStateAndPlan{
			State: state.Disk,
			Plan:  plan.Disk,
		},
		InstanceCount: plan.NumInstances.ValueInt64Pointer(),
		Autoscaling: &common.AutoscalingStateAndPlan{
			State: state.Autoscaling,
			Plan:  plan.Autoscaling,
		},
		EnvironmentID: &common.EnvironmentIDStateAndPlan{
			State: state.EnvironmentID.ValueStringPointer(),
			Plan:  plan.EnvironmentID.ValueStringPointer(),
		},
		NotificationOverride: notificationOverride,
		LogStreamOverride: &common.LogStreamOverrideStateAndPlan{
			State: state.LogStreamOverride,
			Plan:  plan.LogStreamOverride,
		},
	}, common.ServiceTypeBackgroundWorker)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service",
			"Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	bw, err := backgroundWorker.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating background worker",
			"Could not update background worker, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, bw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *backgroundWorkerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backgroundWorker.BackgroundWorkerModel
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

func (r *backgroundWorkerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *backgroundWorkerResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcecommon.RuntimeSourceValidator,
		resourcecommon.ImageTagOrDigestValidator,
		resourcecommon.PreviewGenerationValidator,
	}
}
