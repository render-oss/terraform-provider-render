package resource

import (
	"context"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/project"

	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client  *client.ClientWithResponses
	ownerID string `tfsdk:"owner_id"`
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan project.ProjectModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environmentInput []client.ProjectPOSTEnvironmentInput

	for _, env := range plan.Environments {
		environmentInput = append(environmentInput, client.ProjectPOSTEnvironmentInput{
			Name:                    env.Name.ValueString(),
			ProtectedStatus:         common.From(project.ClientProtectedStatusFromModel(*env)),
			NetworkIsolationEnabled: common.From(env.NetworkIsolated.ValueBool()),
		})
	}

	var createProjectBody = client.CreateProjectJSONRequestBody{
		Name:         plan.Name.ValueString(),
		OwnerId:      r.ownerID,
		Environments: environmentInput,
	}

	var projectResp client.Project
	err := common.Create(func() (*http.Response, error) {
		return r.client.CreateProject(ctx, createProjectBody)
	}, &projectResp)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project", "Could not create project, unexpected error: "+err.Error(),
		)
		return
	}

	environments := make(map[string]*client.Environment)
	for _, envId := range projectResp.EnvironmentIds {
		environment, err := common.GetEnvironmentById(ctx, r.client, envId)
		if err != nil {
			resp.Diagnostics.AddError("Error creating project", "Fetching environment for project failed: "+err.Error())
			return
		}

		var key string
		for k, env := range plan.Environments {
			if env.Name.ValueString() == environment.Name {
				key = k
				break
			}
		}
		environments[key] = environment
	}

	projModel, err := project.ModelForProjectResult(&projectResp, environments)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project", "Could not create project, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, projModel)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state project.ProjectModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectModel, err := project.Read(ctx, r.client, state)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(state.Id.ValueString(), &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project", "Could not read project, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.Set(ctx, projectModel)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan project.ProjectModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state project.ProjectModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var projectUpdateRequest = client.UpdateProjectJSONRequestBody{
		Name: plan.Name.ValueStringPointer(),
	}

	var projectPatchResponse client.Project
	err := common.Update(func() (*http.Response, error) {
		return r.client.UpdateProject(ctx, plan.Id.ValueString(), projectUpdateRequest)
	}, &projectPatchResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating project", "Could not update project, unexpected error : "+err.Error(),
		)
		return
	}

	plan.Name = types.StringValue(projectPatchResponse.Name)

	environments := make(map[string]*project.EnvironmentModel)

	var planKeys []string
	for _, v := range plan.Environments {
		planKeys = append(planKeys, v.Id.ValueString())
	}

	var stateKeys []string
	for _, v := range state.Environments {
		stateKeys = append(stateKeys, v.Id.ValueString())
	}

	_, both, inState := common.XORStringSlices(planKeys, stateKeys)
	for key, env := range plan.Environments {
		if slices.Contains(both, env.Id.ValueString()) {
			// update an existing environment
			envUpdate := client.EnvironmentPATCHInput{
				Name:                    env.Name.ValueStringPointer(),
				ProtectedStatus:         common.From(project.ClientProtectedStatusFromModel(*env)),
				NetworkIsolationEnabled: common.From(env.NetworkIsolated.ValueBool()),
			}

			var environmentResponse client.Environment
			err := common.Update(func() (*http.Response, error) {
				return r.client.UpdateEnvironment(ctx, env.Id.ValueString(), envUpdate)
			}, &environmentResponse)
			environments[key] = common.From(project.ModelForEnvironmentResult(&environmentResponse))

			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating project", "Could not update environment, unexpected error : "+err.Error(),
				)
				return
			}
		} else {
			// If the environment is not in the state, create it
			envCreate := client.EnvironmentPOSTInput{
				ProjectId:               plan.Id.ValueString(),
				Name:                    env.Name.ValueString(),
				ProtectedStatus:         common.From(project.ClientProtectedStatusFromModel(*env)),
				NetworkIsolationEnabled: common.From(env.NetworkIsolated.ValueBool()),
			}

			var environmentResponse client.Environment
			err := common.Create(func() (*http.Response, error) {
				return r.client.CreateEnvironment(ctx, envCreate)
			}, &environmentResponse)
			environments[key] = common.From(project.ModelForEnvironmentResult(&environmentResponse))

			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating project", "Could not create environment, unexpected error : "+err.Error(),
				)
				return
			}
		}
	}

	for _, env := range state.Environments {
		// delete environments we don't want any more
		if slices.Contains(inState, env.Id.ValueString()) {
			err := common.Delete(func() (*http.Response, error) {
				return r.client.DeleteEnvironment(ctx, env.Id.ValueString())
			})

			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating project", "Could not delete environment, unexpected error : "+err.Error(),
				)
				return
			}
		}
	}

	plan.Environments = environments

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var proj project.ProjectModel
	diags := req.State.Get(ctx, &proj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteProject(ctx, proj.Id.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project", "Could not delete project, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
