package resource

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/envgroup"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &envGroupLinkResource{}
	_ resource.ResourceWithConfigure   = &envGroupLinkResource{}
	_ resource.ResourceWithImportState = &envGroupLinkResource{}
)

// NewenvGroupLinkResource is a helper function to simplify the provider implementation.
func NewEnvGroupLinkResource() resource.Resource {
	return &envGroupLinkResource{}
}

// envGroupLinkResource is the resource implementation.
type envGroupLinkResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *envGroupLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *envGroupLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env_group_link"
}

// Schema defines the schema for the resource.
func (r *envGroupLinkResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = EnvGroupLinkResourceSchema(ctx)
}

// Create a new resource.
func (r *envGroupLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan envgroup.EnvGroupLinkModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envGroup, err := r.linkServices(ctx, plan.EnvGroupId.ValueString(), plan.ServiceIds)
	if err != nil {
		resp.Diagnostics.AddError("Unable to add service to environment group", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, envgroup.LinkModelFromClient(&envGroup))
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *envGroupLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id string
	diags := req.State.GetAttribute(ctx, path.Root("env_group_id"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envGroup, err := r.getEnvGroup(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("unable to get environment group", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, envgroup.LinkModelFromClient(envGroup))
	resp.Diagnostics.Append(diags...)
}

func (r *envGroupLinkResource) getEnvGroup(ctx context.Context, id string) (*client.EnvGroup, error) {
	envGroupLinkResp, err := r.client.GetEnvGroupWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}

	if envGroupLinkResp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status: %s", envGroupLinkResp.Status())
	}

	envGroupLink := envGroupLinkResp.JSON200
	return envGroupLink, nil
}

func (r *envGroupLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan envgroup.EnvGroupLinkModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state envgroup.EnvGroupLinkModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var toCreate, toDelete []string
	for _, id := range plan.ServiceIds {
		if !slices.Contains(state.ServiceIds, id) {
			toCreate = append(toCreate, id)
		}
	}

	for _, id := range state.ServiceIds {
		if !slices.Contains(plan.ServiceIds, id) {
			toDelete = append(toDelete, id)
		}
	}

	_, err := r.linkServices(ctx, plan.EnvGroupId.ValueString(), toCreate)
	if err != nil {
		resp.Diagnostics.AddError("Unable to add service to environment group", err.Error())
		return
	}

	err = r.unlinkServices(ctx, plan.EnvGroupId.ValueString(), toDelete)
	if err != nil {
		resp.Diagnostics.AddError("Unable to remove service from environment group", err.Error())
		return
	}

	envGroup, err := r.getEnvGroup(ctx, plan.EnvGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to get environment group", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, envgroup.LinkModelFromClient(envGroup))
	resp.Diagnostics.Append(diags...)
}

func (r *envGroupLinkResource) linkServices(ctx context.Context, envGroupID string, serviceIDs []string) (client.EnvGroup, error) {
	var envGroup client.EnvGroup
	for _, serviceId := range serviceIDs {
		err := common.Create(func() (*http.Response, error) {
			return r.client.LinkServiceToEnvGroup(ctx, envGroupID, serviceId)
		}, &envGroup)
		if err != nil {
			return client.EnvGroup{}, err
		}
	}

	return envGroup, nil
}

func (r *envGroupLinkResource) unlinkServices(ctx context.Context, envGroupID string, serviceIDs []string) error {
	for _, serviceId := range serviceIDs {
		err := common.Delete(func() (*http.Response, error) {
			return r.client.UnlinkServiceFromEnvGroup(ctx, envGroupID, serviceId)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *envGroupLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state envgroup.EnvGroupLinkModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.unlinkServices(ctx, state.EnvGroupId.ValueString(), state.ServiceIds)
	if err != nil {
		resp.Diagnostics.AddError("Unable to remove service from environment group", err.Error())
		return
	}
}

func (r *envGroupLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("env_group_id"), req, resp)
}
