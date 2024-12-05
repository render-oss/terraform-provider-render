package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/staticsite"
	"terraform-provider-render/internal/provider/staticsite/resource/internal"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &staticSiteResource{}
	_ resource.ResourceWithConfigure   = &staticSiteResource{}
	_ resource.ResourceWithImportState = &staticSiteResource{}
)

// NewStaticSiteResource is a helper function to simplify the provider implementation.
func NewStaticSiteResource() resource.Resource {
	return &staticSiteResource{}
}

// staticSiteResource is the resource implementation.
type staticSiteResource struct {
	client                       *client.ClientWithResponses
	ownerID                      string
	skipDeployAfterServiceUpdate bool
}

// Configure adds the provider configured Client to the resource.
func (r *staticSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
	r.skipDeployAfterServiceUpdate = data.SkipDeployAfterServiceUpdate
}

// Metadata returns the resource type name.
func (r *staticSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site"
}

// Schema defines the schema for the resource.
func (r *staticSiteResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *staticSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan staticsite.StaticSiteModel
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

	service, err := common.CreateService(ctx, r.client, common.CreateServiceReq{
		Service:              serviceDetails,
		CustomDomains:        common.CustomDomainModelsToClientCustomDomains(plan.CustomDomains),
		EnvironmentID:        plan.EnvironmentID.ValueStringPointer(),
		NotificationOverride: plan.NotificationOverride,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating static site", err.Error(),
		)
		return
	}

	staticSite, err := common.WrapStaticSite(ctx, r.client, service.Service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating static site", err.Error(),
		)
		return
	}

	staticSiteModel, err := staticsite.ModelForServiceResult(staticSite, plan, diags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating static site", "Could not create static site, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, staticSiteModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *staticSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state staticsite.StaticSiteModel
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
			"Error reading service", "Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	staticSite, err := common.WrapStaticSite(ctx, r.client, service.Service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service", "Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	staticSiteModel, err := staticsite.ModelForServiceResult(staticSite, state, diags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service", "Could not read service, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, staticSiteModel)
	resp.Diagnostics.Append(diags...)
}

func (r *staticSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan staticsite.StaticSiteModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state staticsite.StaticSiteModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.UpdateServiceRequestFromModel(ctx, plan)
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

	wrappedService, err := common.UpdateStaticSite(ctx, r.client, r.skipDeployAfterServiceUpdate, common.UpdateStaticSiteReq{
		ServiceID: plan.Id.ValueString(),
		Service:   serviceDetails,
		CustomDomains: common.CustomDomainStateAndPlan{
			State: state.CustomDomains,
			Plan:  plan.CustomDomains,
		},
		EnvVars:              evs,
		Headers:              common.ModelToClientHeaderInput(plan.Headers),
		NotificationOverride: notificationOverride,
		Routes:               common.RouteModelToClientRoutePutInput(plan.Routes),
		EnvironmentID: &common.EnvironmentIDStateAndPlan{
			State: state.EnvironmentID.ValueStringPointer(),
			Plan:  plan.EnvironmentID.ValueStringPointer(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service", "Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	model, err := staticsite.ModelForServiceResult(wrappedService, plan, diags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service", "Could not update service, unexpected error: "+err.Error(),
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service", "Could not update service, unexpected error: headers not found",
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *staticSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state staticsite.StaticSiteModel
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
			"Error deleting service", "Could not delete service, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *staticSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
