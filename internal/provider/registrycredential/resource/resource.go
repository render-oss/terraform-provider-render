package resource

import (
	"context"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &registryCredentialResource{}
	_ resource.ResourceWithConfigure   = &registryCredentialResource{}
	_ resource.ResourceWithImportState = &registryCredentialResource{}
)

// NewregistryCredentialResource is a helper function to simplify the provider implementation.
func NewRegistryCredentialResource() resource.Resource {
	return &registryCredentialResource{}
}

// registryCredentialResource is the resource implementation.
type registryCredentialResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *registryCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *registryCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_credential"
}

// Schema defines the schema for the resource.
func (r *registryCredentialResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = RegistryCredentialResourceSchema(ctx)
}

// Create a new resource.
func (r *registryCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan RegistryCredentialModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var registry client.RegistryCredentialRegistry
	switch plan.Registry {
	case types.StringValue("GITHUB"):
		registry = client.GITHUB
	case types.StringValue("GITLAB"):
		registry = client.GITLAB
	case types.StringValue("DOCKER"):
		registry = client.DOCKER
	default:
		resp.Diagnostics.AddError(
			"Error creating registry credential",
			"Invalid registry type: "+plan.Registry.ValueString(),
		)
		return
	}

	requestBody := client.CreateRegistryCredentialJSONRequestBody{
		AuthToken: plan.AuthToken.ValueString(),
		Name:      plan.Name.ValueString(),
		OwnerId:   r.ownerID,
		Registry:  registry,
		Username:  plan.Username.ValueString(),
	}
	response, err := r.client.CreateRegistryCredentialWithResponse(ctx, requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating registry credential",
			"Could not create registry credential, unexpected error: "+err.Error(),
		)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error creating registry credential",
			"Could not create registry credential, unexpected status code: "+strconv.Itoa(response.StatusCode()),
		)
		return
	}

	regCred := response.JSON200
	plan.Id = types.StringValue(regCred.Id)
	plan.Name = types.StringValue(regCred.Name)
	plan.Username = types.StringValue(regCred.Username)
	plan.Registry = types.StringValue(string(regCred.Registry))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *registryCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state RegistryCredentialModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.RetrieveRegistryCredentialWithResponse(ctx, state.Id.ValueString())
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(state.Id.ValueString(), &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting registry credential",
			"Could not get registry credential for credential ID: "+state.Id.ValueString()+": "+err.Error(),
		)
		return
	}

	if response.StatusCode() == 404 {
		state.Name = types.StringNull()
		state.Username = types.StringNull()
		state.Registry = types.StringNull()
		state.Id = types.StringNull()

		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)

		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error getting registry credential",
			"Could not get registry credential for credential ID: "+state.Id.ValueString()+": unexpected status code: "+strconv.Itoa(response.StatusCode()),
		)

		return
	}

	registryCredential := response.JSON200

	if registryCredential.Id == "" {
		resp.Diagnostics.AddError(
			"Error getting registry credential",
			"Could not get registry credential for credential ID: "+state.Id.ValueString()+": unexpected empty ID",
		)

		return
	}
	// Update state with registry credential data
	state.Name = types.StringValue(registryCredential.Name)
	state.Username = types.StringValue(registryCredential.Username)
	state.Registry = types.StringValue(string(registryCredential.Registry))
	state.Id = types.StringValue(registryCredential.Id)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *registryCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan RegistryCredentialModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := client.UpdateRegistryCredentialJSONRequestBody{
		AuthToken: plan.AuthToken.ValueString(),
		Name:      plan.Name.ValueString(),
		Registry:  client.RegistryCredentialRegistry(plan.Registry.ValueString()),
		Username:  plan.Username.ValueString(),
	}

	if plan.Id.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Error updating registry credential",
			"ID is required to update registry credential",
		)
		return
	}

	response, err := r.client.UpdateRegistryCredentialWithResponse(ctx, plan.Id.ValueString(), requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating registry credential",
			"Could not update registry credential, unexpected error: "+err.Error(),
		)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error updating registry credential",
			"Could not update registry credential, unexpected status code: "+strconv.Itoa(response.StatusCode()),
		)
		return
	}

	regCred := response.JSON200
	plan.Id = types.StringValue(regCred.Id)
	plan.Name = types.StringValue(regCred.Name)
	plan.Username = types.StringValue(regCred.Username)
	plan.Registry = types.StringValue(string(regCred.Registry))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *registryCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state RegistryCredentialModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteRegistryCredential(ctx, state.Id.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting registry credential",
			"Could not delete registry credential, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *registryCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// todo
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
