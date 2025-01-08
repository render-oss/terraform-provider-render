package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/keyvalue"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/keyvalue/resource/internal"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &keyvalueResource{}
	_ resource.ResourceWithConfigure        = &keyvalueResource{}
	_ resource.ResourceWithImportState      = &keyvalueResource{}
	_ resource.ResourceWithConfigValidators = &keyvalueResource{}
)

// NewKeyValueResource is a helper function to simplify the provider implementation.
func NewKeyValueResource() resource.Resource {
	return &keyvalueResource{}
}

// keyvalueResource is the resource implementation.
type keyvalueResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *keyvalueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *keyvalueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keyvalue"
}

// Schema defines the schema for the resource.
func (r *keyvalueResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *keyvalueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan keyvalue.KeyValueModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceDetails, err := internal.CreateKeyValueRequestFromModel(r.ownerID, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service", "Could not create service, unexpected error: "+err.Error(),
		)
		return
	}

	var model client.KeyValue
	err = common.Create(func() (*http.Response, error) {
		return r.client.CreateKeyValue(ctx, serviceDetails)
	}, &model)
	if err != nil {
		resp.Diagnostics.AddError("Error creating service", err.Error())
		return
	}

	var connectionInfo client.KeyValueConnectionInfo
	if err = common.Get(func() (*http.Response, error) {
		return r.client.RetrieveKeyValueConnectionInfo(ctx, model.Id)
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get keyvalue connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.UpdateLogStreamOverride(
		ctx,
		r.client,
		model.Id,
		&common.LogStreamOverrideStateAndPlan{
			Plan: plan.LogStreamOverride,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("unable to create log stream overrides", err.Error())
		return
	}
	keyvalueModel := keyvalue.ModelForKeyValueResult(&model, &plan, &connectionInfo, logStreamOverrides, resp.Diagnostics)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, keyvalueModel)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *keyvalueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state keyvalue.KeyValueModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var clientKeyValue client.KeyValue
	err := common.Get(func() (*http.Response, error) {
		return r.client.RetrieveKeyValue(ctx, state.Id.ValueString())
	}, &clientKeyValue)
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

	var connectionInfo client.KeyValueConnectionInfo
	if err = common.Get(func() (*http.Response, error) {
		return r.client.RetrieveKeyValueConnectionInfo(ctx, clientKeyValue.Id)
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get keyvalue connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.GetLogStreamOverrides(ctx, r.client, clientKeyValue.Id)
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream overrides", err.Error())
		return
	}

	keyvalueModel := keyvalue.ModelForKeyValueResult(&clientKeyValue, &state, &connectionInfo, logStreamOverrides, resp.Diagnostics)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, keyvalueModel)
	resp.Diagnostics.Append(diags...)
}

func (r *keyvalueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan keyvalue.KeyValueModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state keyvalue.KeyValueModel
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

	var keyvalueResponse client.KeyValue
	err = common.Update(func() (*http.Response, error) {
		return r.client.UpdateKeyValue(ctx, plan.Id.ValueString(), serviceDetails)
	}, &keyvalueResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating keyvalue", err.Error(),
		)
		return
	}

	var connectionInfo client.KeyValueConnectionInfo
	if err = common.Get(func() (*http.Response, error) {
		return r.client.RetrieveKeyValueConnectionInfo(ctx, keyvalueResponse.Id)
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get keyvalue connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.UpdateLogStreamOverride(
		ctx,
		r.client,
		keyvalueResponse.Id,
		&common.LogStreamOverrideStateAndPlan{
			Plan:  plan.LogStreamOverride,
			State: state.LogStreamOverride,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream overrides", err.Error())
		return
	}

	keyvalueModel := keyvalue.ModelForKeyValueResult(&keyvalueResponse, &plan, &connectionInfo, logStreamOverrides, resp.Diagnostics)

	envID, err := common.UpdateEnvironmentID(ctx, r.client, keyvalueModel.Id.ValueString(), &common.EnvironmentIDStateAndPlan{
		State: state.EnvironmentID.ValueStringPointer(),
		Plan:  plan.EnvironmentID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating environment ID", err.Error(),
		)
		return
	}

	keyvalueModel.EnvironmentID = types.StringPointerValue(envID)

	diags = resp.State.Set(ctx, keyvalueModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *keyvalueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state keyvalue.KeyValueModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteKeyValue(ctx, state.Id.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting keyvalue",
			err.Error(),
		)
		return
	}
}

func (r *keyvalueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *keyvalueResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}
