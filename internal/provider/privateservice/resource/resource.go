package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/privateservice"
	"terraform-provider-render/internal/provider/privateservice/resource/internal"
	rendertypes "terraform-provider-render/internal/provider/types"
	resourcecommon "terraform-provider-render/internal/provider/types/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &privateServiceResource{}
	_ resource.ResourceWithConfigure        = &privateServiceResource{}
	_ resource.ResourceWithImportState      = &privateServiceResource{}
	_ resource.ResourceWithConfigValidators = &privateServiceResource{}
)

// NewPrivateServiceResource is a helper function to simplify the provider implementation.
func NewPrivateServiceResource() resource.Resource {
	return &privateServiceResource{}
}

// privateServiceResource is the resource implementation.
type privateServiceResource struct {
	client                  *client.ClientWithResponses
	ownerID                 string
	poller                  *common.Poller
	waitForDeployCompletion bool
}

// Configure adds the provider configured Client to the resource.
func (r *privateServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
	r.poller = data.Poller
	r.waitForDeployCompletion = data.WaitForDeployCompletion
}

// Metadata returns the resource type name.
func (r *privateServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_service"
}

// Schema defines the schema for the resource.
func (r *privateServiceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *privateServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var shouldWaitForServiceCompletion = r.waitForDeployCompletion
	var plan privateservice.PrivateServiceModel
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
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service", "Could not create service, unexpected error: "+err.Error(),
		)
		// We don't return early here because we want to add the service to the state
		// if it was created. Even if there was an error during creation, the service
		// may be in a partial created state.
		shouldWaitForServiceCompletion = false
	}

	service, err := common.GetWrappedServiceByName(ctx, r.client, r.ownerID, plan.Name.ValueString(), client.PrivateService)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating web service", "Could not find service, unexpected error: "+err.Error(),
		)
		return
	}

	model, err := privateservice.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating web service", "Could not create web service, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, *model)
	resp.Diagnostics.Append(diags...)

	if !shouldWaitForServiceCompletion {
		return
	}

	err = common.WaitForService(ctx, r.poller, r.client, service.Id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating private service",
			"Service never started: "+err.Error(),
		)
		return
	}
}

// Read resource information.
func (r *privateServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privateservice.PrivateServiceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := common.GetWrappedService(ctx, r.client, state.Id.ValueString())
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(state.Id.ValueString(), &diags)
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

	privateServiceModel, err := privateservice.ModelForServiceResult(service, state, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service",
			"Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, privateServiceModel)
	resp.Diagnostics.Append(diags...)
}

func (r *privateServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan privateservice.PrivateServiceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state privateservice.PrivateServiceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.UpdateServiceRequestFromModel(ctx, plan, r.ownerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service",
			"Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	evs, err := common.EnvVarsToClient(plan.EnvVars)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service",
			"Could not process env vars, unexpected error: "+err.Error(),
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
	}, common.ServiceTypePrivateService)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service",
			"Could not update service, unexpected error: "+err.Error(),
		)
		return
	}
	privateserviceModel, err := privateservice.ModelForServiceResult(service, plan, resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service",
			"Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, privateserviceModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privateServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state privateservice.PrivateServiceModel
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
			"Could not delete service, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *privateServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *privateServiceResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcecommon.RuntimeSourceValidator,
		resourcecommon.ImageTagOrDigestValidator,
		resourcecommon.PreviewGenerationValidator,
	}
}
