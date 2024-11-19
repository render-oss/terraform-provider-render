package resource

import (
	"context"
	"net/http"
	"time"

	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/postgres"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	clientpostgres "terraform-provider-render/internal/client/postgres"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &postgresResource{}
	_ resource.ResourceWithConfigure   = &postgresResource{}
	_ resource.ResourceWithImportState = &postgresResource{}
)

// NewPostgresResource is a helper function to simplify the provider implementation.
func NewPostgresResource() resource.Resource {
	return &postgresResource{}
}

// postgresResource is the resource implementation.
type postgresResource struct {
	client  *client.ClientWithResponses
	ownerID string
	poller  *common.Poller
}

// Configure adds the provider configured Client to the resource.
func (r *postgresResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
	r.poller = data.Poller
}

// Metadata returns the resource type name.
func (r *postgresResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres"
}

// Schema defines the schema for the resource.
func (r *postgresResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PostgresResourceSchema(ctx)
}

// Create a new resource.
func (r *postgresResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan postgres.PostgresModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pg client.Postgres

	ipAllowList, err := common.ClientFromIPAllowList(plan.IPAllowList)
	if err != nil {
		resp.Diagnostics.AddError("unable to parse ip allow list", err.Error())
		return
	}

	err = common.Create(func() (*http.Response, error) {
		return r.client.CreatePostgres(ctx, client.PostgresPOSTInput{
			DatabaseName:           plan.DatabaseName.ValueStringPointer(),
			DatabaseUser:           plan.DatabaseUser.ValueStringPointer(),
			DatadogAPIKey:          plan.DatadogAPIKey.ValueStringPointer(),
			EnableHighAvailability: plan.HighAvailabilityEnabled.ValueBoolPointer(),
			EnvironmentId:          plan.EnvironmentID.ValueStringPointer(),
			IpAllowList:            common.From(ipAllowList),
			Plan:                   clientpostgres.PostgresPlans(plan.Plan.ValueString()),
			ReadReplicas:           common.From(postgres.ReadReplicaInputFromModel(plan.ReadReplicas)),
			Region:                 plan.Region.ValueStringPointer(),
			Version:                client.PostgresVersion(plan.Version.ValueString()),
			Name:                   plan.Name.ValueString(),
			OwnerId:                r.ownerID,
			DiskSizeGB:             common.ValueAsIntPointer(plan.DiskSizeGB),
		})
	}, &pg)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create postgres database", err.Error())
		return
	}

	// Poll for postgres to be ready
	err = r.poller.Poll(ctx, func() (bool, error) {
		var polledPG client.Postgres
		err := common.Get(func() (*http.Response, error) {
			return r.client.RetrievePostgres(ctx, pg.Id)
		}, &polledPG)
		if err != nil {
			return false, err
		}

		return polledPG.Status == client.DatabaseStatusAvailable, nil
	}, 15*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("postgres never became available", err.Error())
		return
	}

	var connectionInfo client.PostgresConnectionInfo
	if err = common.Get(func() (*http.Response, error) {
		return r.client.RetrievePostgresConnectionInfo(ctx, pg.Id)
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get postgres connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.UpdateLogStreamOverride(
		ctx,
		r.client,
		pg.Id,
		&common.LogStreamOverrideStateAndPlan{
			Plan: plan.LogStreamOverride,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("unable to create log stream overrides", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, postgres.ModelFromClient(&pg, &connectionInfo, logStreamOverrides, plan, resp.Diagnostics))
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *postgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	id, ok := common.IDFromState(ctx, req.State, resp.Diagnostics)
	if !ok {
		return
	}

	var state postgres.PostgresModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pg client.Postgres

	err := common.Get(func() (*http.Response, error) {
		return r.client.RetrievePostgres(ctx, id)
	}, &pg)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(id, &diags)
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("unable to get postgres", err.Error())
		return
	}

	var connectionInfo client.PostgresConnectionInfo
	if err = common.Get(func() (*http.Response, error) {
		return r.client.RetrievePostgresConnectionInfo(ctx, id)
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get postgres connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.GetLogStreamOverrides(ctx, r.client, id)
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream overrides", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, postgres.ModelFromClient(&pg, &connectionInfo, logStreamOverrides, state, resp.Diagnostics))
	resp.Diagnostics.Append(diags...)
}

func (r *postgresResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan postgres.PostgresModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state postgres.PostgresModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipAllowList, err := common.ClientFromIPAllowList(plan.IPAllowList)
	if err != nil {
		resp.Diagnostics.AddError("unable to parse ip allow list", err.Error())
		return
	}

	var pg client.Postgres

	err = common.Update(func() (*http.Response, error) {
		return r.client.UpdatePostgres(ctx, plan.ID.ValueString(), client.PostgresPATCHInput{
			EnableHighAvailability: plan.HighAvailabilityEnabled.ValueBoolPointer(),
			IpAllowList:            common.From(ipAllowList),
			Name:                   plan.Name.ValueStringPointer(),
			Plan:                   common.From(clientpostgres.PostgresPlans(plan.Plan.ValueString())),
			DatadogAPIKey:          plan.DatadogAPIKey.ValueStringPointer(),
			ReadReplicas:           common.From(postgres.ReadReplicaInputFromModel(plan.ReadReplicas)),
			DiskSizeGB:             common.ValueAsIntPointer(plan.DiskSizeGB),
		})
	}, &pg)
	if err != nil {
		resp.Diagnostics.AddError("unable to update postgres", err.Error())
		return
	}

	envID, err := common.UpdateEnvironmentID(ctx, r.client, pg.Id, &common.EnvironmentIDStateAndPlan{
		State: state.EnvironmentID.ValueStringPointer(),
		Plan:  plan.EnvironmentID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("unable to update environment ID", err.Error())
		return
	}
	pg.EnvironmentId = envID

	var connectionInfo client.PostgresConnectionInfo
	if err = common.Get(func() (*http.Response, error) {
		return r.client.RetrievePostgresConnectionInfo(ctx, plan.ID.ValueString())
	}, &connectionInfo); err != nil {
		resp.Diagnostics.AddError("unable to get postgres connection info", err.Error())
		return
	}

	logStreamOverrides, err := common.UpdateLogStreamOverride(
		ctx,
		r.client,
		plan.ID.ValueString(),
		&common.LogStreamOverrideStateAndPlan{
			Plan:  plan.LogStreamOverride,
			State: state.LogStreamOverride,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream overrides", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, postgres.ModelFromClient(&pg, &connectionInfo, logStreamOverrides, plan, resp.Diagnostics))
	resp.Diagnostics.Append(diags...)
}

func (r *postgresResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	id, ok := common.IDFromState(ctx, req.State, resp.Diagnostics)
	if !ok {
		return
	}

	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeletePostgres(ctx, id)
	})
	if err != nil {
		resp.Diagnostics.AddError("unable to delete postgres", err.Error())
		return
	}
}

func (r *postgresResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
