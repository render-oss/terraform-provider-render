package resource

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/dedicatedip"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// dedicatedIPProvisioningTimeout bounds Create polling. EIP allocation is
// typically 30-90s; 15m matches the postgres resource's tolerance.
const dedicatedIPProvisioningTimeout = 15 * time.Minute

var (
	_ resource.Resource                = &dedicatedIPResource{}
	_ resource.ResourceWithConfigure   = &dedicatedIPResource{}
	_ resource.ResourceWithImportState = &dedicatedIPResource{}
)

func NewDedicatedIPResource() resource.Resource {
	return &dedicatedIPResource{}
}

type dedicatedIPResource struct {
	client  *client.ClientWithResponses
	ownerID string
	poller  *common.Poller
}

func (r *dedicatedIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}
	r.client = data.Client
	r.ownerID = data.OwnerID
	r.poller = data.Poller
}

func (r *dedicatedIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dedicated_ip"
}

func (r *dedicatedIPResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

func (r *dedicatedIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dedicatedip.Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.DedicatedIPPOST{
		Name:    plan.Name.ValueString(),
		OwnerId: r.ownerID,
		Region:  client.Region(plan.Region.ValueString()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		desc := plan.Description.ValueString()
		body.Description = &desc
	}
	envIDs := dedicatedip.EnvironmentIDsFromPlan(plan.EnvironmentIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	body.EnvironmentIds = &envIDs

	var created client.DedicatedIP
	if err := common.Create(func() (*http.Response, error) {
		return r.client.CreateDedicatedIp(ctx, body)
	}, &created); err != nil {
		resp.Diagnostics.AddError("Error creating dedicated IP", err.Error())
		return
	}

	// Provisioning is asynchronous. POST returns immediately with
	// status=CREATING and ips=[]; status transitions to RUNNING once the
	// EgressGatewaySet's Elastic IPs are allocated. Poll until then so
	// computed attributes (ips, status) are populated when apply returns.
	final := created
	if err := r.poller.Poll(ctx, func() (bool, error) {
		var p client.DedicatedIP
		if err := common.Get(func() (*http.Response, error) {
			return r.client.RetrieveDedicatedIp(ctx, created.Id)
		}, &p); err != nil {
			return false, err
		}
		switch p.Status {
		case client.RUNNING:
			final = p
			return true, nil
		case client.FAILED:
			return false, fmt.Errorf("dedicated IP %s entered FAILED status during provisioning", created.Id)
		case client.DELETING, client.DELETED:
			return false, fmt.Errorf("dedicated IP %s was deleted during provisioning", created.Id)
		}
		return false, nil
	}, dedicatedIPProvisioningTimeout); err != nil {
		resp.Diagnostics.AddError("dedicated IP never became RUNNING", err.Error())
		return
	}

	state := dedicatedip.ModelFromClient(&final, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *dedicatedIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dedicatedip.Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var fetched client.DedicatedIP
	err := common.Get(func() (*http.Response, error) {
		return r.client.RetrieveDedicatedIp(ctx, state.ID.ValueString())
	}, &fetched)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(state.ID.ValueString(), &resp.Diagnostics)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading dedicated IP", err.Error())
		return
	}

	newState := dedicatedip.ModelFromClient(&fetched, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *dedicatedIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dedicatedip.Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state dedicatedip.Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.DedicatedIPPATCH{}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		body.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		body.Description = &v
	}
	if !plan.EnvironmentIDs.Equal(state.EnvironmentIDs) {
		envIDs := dedicatedip.EnvironmentIDsFromPlan(plan.EnvironmentIDs, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		body.EnvironmentIds = &envIDs
	}

	var updated client.DedicatedIP
	if err := common.Update(func() (*http.Response, error) {
		return r.client.UpdateDedicatedIp(ctx, plan.ID.ValueString(), body)
	}, &updated); err != nil {
		resp.Diagnostics.AddError("Error updating dedicated IP", err.Error())
		return
	}

	newState := dedicatedip.ModelFromClient(&updated, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *dedicatedIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dedicatedip.Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteDedicatedIp(ctx, state.ID.ValueString())
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting dedicated IP", err.Error())
		return
	}
}

func (r *dedicatedIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
