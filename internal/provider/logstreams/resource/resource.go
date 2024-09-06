package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"terraform-provider-render/internal/client/logs"
	"terraform-provider-render/internal/provider/logstreams"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &logStreamSettingResource{}
	_ resource.ResourceWithConfigure = &logStreamSettingResource{}
)

// NewLogStreamSettingResource is a helper function to simplify the provider implementation.
func NewLogStreamSettingResource() resource.Resource {
	return &logStreamSettingResource{}
}

// logStreamSettingResource is the resource implementation.
type logStreamSettingResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *logStreamSettingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *logStreamSettingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_stream"
}

// Schema defines the schema for the resource.
func (r *logStreamSettingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *logStreamSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan logstreams.LogStreamSettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var logStream logs.OwnerLogStreamSetting

	put := r.logStreamPut(plan)

	err := common.Update(func() (*http.Response, error) {
		return r.client.UpdateOwnerLogStream(ctx, r.ownerID, put)
	}, &logStream)
	if err != nil {
		resp.Diagnostics.AddError("unable to create log stream settings", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, logstreams.LogStreamFromClient(&logStream, plan, diags))
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *logStreamSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var settings logstreams.LogStreamSettingModel
	diags := req.State.Get(ctx, &settings)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	var logStream logs.OwnerLogStreamSetting

	err := common.Get(func() (*http.Response, error) {
		return r.client.GetOwnerLogStream(ctx, r.ownerID)
	}, &logStream)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(r.ownerID, &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("unable to get log stream settings", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, logstreams.LogStreamFromClient(&logStream, settings, diags))
	resp.Diagnostics.Append(diags...)
}

func (r *logStreamSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan logstreams.LogStreamSettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var logStream logs.OwnerLogStreamSetting

	put := r.logStreamPut(plan)

	err := common.Update(func() (*http.Response, error) {
		return r.client.UpdateOwnerLogStream(ctx, r.ownerID, put)
	}, &logStream)
	if err != nil {
		resp.Diagnostics.AddError("unable to create log stream settings", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, logstreams.LogStreamFromClient(&logStream, plan, diags))
	resp.Diagnostics.Append(diags...)
}

func (r *logStreamSettingResource) logStreamPut(plan logstreams.LogStreamSettingModel) client.UpdateOwnerLogStreamJSONRequestBody {
	preview := logs.LogStreamPreviewSettingSend
	if plan.Preview.ValueString() == string(logs.LogStreamPreviewSettingDrop) {
		preview = logs.LogStreamPreviewSettingDrop
	}

	return client.UpdateOwnerLogStreamJSONRequestBody{
		Preview:  &preview,
		Endpoint: common.From(plan.Endpoint.ValueString()),
		Token:    common.From(plan.Token.ValueString()),
	}
}

func (r *logStreamSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteOwnerLogStream(ctx, r.ownerID)
	})
	if err != nil {
		resp.Diagnostics.AddError("unable to delete log stream settings", err.Error())
		return
	}
}
