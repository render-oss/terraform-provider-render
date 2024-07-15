package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/envgroup"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &envGroupResource{}
	_ resource.ResourceWithConfigure   = &envGroupResource{}
	_ resource.ResourceWithImportState = &envGroupResource{}
)

// NewenvGroupResource is a helper function to simplify the provider implementation.
func NewEnvGroupResource() resource.Resource {
	return &envGroupResource{}
}

// envGroupResource is the resource implementation.
type envGroupResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *envGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *envGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_group"
}

// Schema defines the schema for the resource.
func (r *envGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = EnvGroupResourceSchema(ctx)
}

// Create a new resource.
func (r *envGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan envgroup.EnvGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envVars, err := common.EnvVarsToClient(plan.EnvVars)
	if err != nil {
		resp.Diagnostics.AddError("invalid env vars", err.Error())
		return
	}

	secretFiles := common.SecretFilesToClient(plan.SecretFiles)

	var envGroup client.EnvGroup
	err = common.Create(func() (*http.Response, error) {
		return r.client.CreateEnvGroup(ctx, client.CreateEnvGroupJSONRequestBody{
			Name:          plan.Name.ValueString(),
			OwnerId:       r.ownerID,
			EnvVars:       envVars,
			EnvironmentId: plan.EnvironmentID.ValueStringPointer(),
			SecretFiles:   &secretFiles,
		})
	}, &envGroup)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create environment group", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, envgroup.ModelFromClient(&envGroup, plan.EnvVars))
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *envGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	id, ok := common.IDFromState(ctx, req.State, resp.Diagnostics)
	if !ok {
		return
	}

	evs, ok := common.EnvVarsFromState(ctx, req.State, resp.Diagnostics)
	if !ok {
		return
	}

	state := &envgroup.EnvGroupModel{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envGroup := client.EnvGroup{}
	err := common.Get(func() (*http.Response, error) {
		return r.client.RetrieveEnvGroup(ctx, id)
	}, &envGroup)

	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(id, &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("unable to get environment group", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, envgroup.ModelFromClient(&envGroup, evs))
	resp.Diagnostics.Append(diags...)
}

func (r *envGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan envgroup.EnvGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state envgroup.EnvGroupModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envGroupID := state.Id.ValueString()

	for k, v := range plan.EnvVars {
		existingVal, exists := state.EnvVars[k]
		if !exists || (!v.GenerateValue.ValueBool() && existingVal.Value != v.Value) {
			err := r.updateEnvVar(ctx, envGroupID, k, v)
			if err != nil {
				resp.Diagnostics.AddError("unable to create or update env var: "+k, err.Error())
				return
			}
		}
	}

	for k := range state.EnvVars {
		if _, ok := plan.EnvVars[k]; !ok {
			err := r.deleteEnvVar(ctx, envGroupID, k)
			if err != nil {
				resp.Diagnostics.AddError("unable to remove env var: "+k, err.Error())
				return
			}
		}
	}

	for k, v := range plan.SecretFiles {
		existingVal, exists := state.SecretFiles[k]
		if !exists || existingVal.Content != v.Content {
			err := r.updateSecretFile(ctx, envGroupID, k, v.Content)
			if err != nil {
				resp.Diagnostics.AddError("unable to create or update secret file: "+k, err.Error())
				return
			}
		}
	}

	for k := range state.SecretFiles {
		if _, ok := plan.SecretFiles[k]; !ok {
			err := r.deleteSecretFile(ctx, envGroupID, k)
			if err != nil {
				resp.Diagnostics.AddError("unable to remove secret file: "+k, err.Error())
				return
			}
		}
	}

	_, err := common.UpdateEnvironmentID(ctx, r.client, envGroupID, &common.EnvironmentIDStateAndPlan{
		State: state.EnvironmentID.ValueStringPointer(),
		Plan:  plan.EnvironmentID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("unable to update environment ID", err.Error())
		return
	}

	envGroup := &client.EnvGroup{}
	if !state.Name.Equal(plan.Name) {
		envGroup, err = r.updateEnvGroupName(ctx, envGroupID, plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("unable to update environment group", err.Error())
			return
		}
	} else {
		err := common.Get(func() (*http.Response, error) {
			return r.client.RetrieveEnvGroup(ctx, envGroupID)
		}, envGroup)
		if err != nil {
			resp.Diagnostics.AddError("unable to get environment group", err.Error())
			return
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, envgroup.ModelFromClient(envGroup, nil))
	resp.Diagnostics.Append(diags...)
}

func (r *envGroupResource) updateEnvVar(ctx context.Context, envGroupID, key string, value common.EnvVarModel) error {
	body, err := common.EnvVarAddUpdateToClient(key, value)
	if err != nil {
		return err
	}

	resp, err := r.client.UpdateEnvGroupEnvVar(ctx, envGroupID, key, *body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (r *envGroupResource) deleteEnvVar(ctx context.Context, envGroupID string, key string) error {
	resp, err := r.client.DeleteEnvGroupEnvVar(ctx, envGroupID, key)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (r *envGroupResource) updateSecretFile(ctx context.Context, envGroupID, key, content string) error {
	resp, err := r.client.UpdateEnvGroupSecretFile(ctx, envGroupID, key, client.UpdateEnvGroupSecretFileJSONRequestBody{
		Content: &content,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (r *envGroupResource) deleteSecretFile(ctx context.Context, envGroupID string, key string) error {
	resp, err := r.client.DeleteEnvGroupSecretFile(ctx, envGroupID, key)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (r *envGroupResource) updateEnvGroupName(ctx context.Context, envGroupID, name string) (*client.EnvGroup, error) {
	envGroupResp, err := r.client.UpdateEnvGroupWithResponse(ctx, envGroupID, client.UpdateEnvGroupJSONRequestBody{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	if envGroupResp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status: %s", envGroupResp.Status())
	}

	envGroup := envGroupResp.JSON200
	return envGroup, nil
}

func (r *envGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	id, ok := common.IDFromState(ctx, req.State, resp.Diagnostics)
	if !ok {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteEnvGroup(ctx, id)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting environment group",
			"Could not delete environment group, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *envGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
