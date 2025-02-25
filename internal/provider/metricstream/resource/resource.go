package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"terraform-provider-render/internal/client/metrics"
	"terraform-provider-render/internal/provider/metricstream"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &metricStreamSettingResource{}
	_ resource.ResourceWithConfigure = &metricStreamSettingResource{}
)

// NewMetricsStreamSettingResource is a helper function to simplify the provider implementation.
func NewMetricsStreamSettingResource() resource.Resource {
	return &metricStreamSettingResource{}
}

// metricStreamSettingResource is the resource implementation.
type metricStreamSettingResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *metricStreamSettingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *metricStreamSettingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics_stream"
}

// Schema defines the schema for the resource.
func (r *metricStreamSettingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

// Create a new resource.
func (r *metricStreamSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricstream.MetricStreamSettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var metricStream metrics.MetricsStream

	put := r.metricStreamPut(plan)

	err := common.Update(func() (*http.Response, error) {
		return r.client.UpsertOwnerMetricsStream(ctx, r.ownerID, put)
	}, &metricStream)
	if err != nil {
		resp.Diagnostics.AddError("unable to create metric stream settings", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, metricstream.MetricStreamFromClient(&metricStream, plan, diags))
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *metricStreamSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var settings metricstream.MetricStreamSettingModel
	diags := req.State.Get(ctx, &settings)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	var metricStream metrics.MetricsStream

	err := common.Get(func() (*http.Response, error) {
		return r.client.GetOwnerMetricsStream(ctx, r.ownerID)
	}, &metricStream)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(r.ownerID, &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("unable to get metric stream settings", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, metricstream.MetricStreamFromClient(&metricStream, settings, diags))
	resp.Diagnostics.Append(diags...)
}

func (r *metricStreamSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan metricstream.MetricStreamSettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var metricStream metrics.MetricsStream

	put := r.metricStreamPut(plan)

	err := common.Update(func() (*http.Response, error) {
		return r.client.UpsertOwnerMetricsStream(ctx, r.ownerID, put)
	}, &metricStream)
	if err != nil {
		resp.Diagnostics.AddError("unable to create metric stream settings", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, metricstream.MetricStreamFromClient(&metricStream, plan, diags))
	resp.Diagnostics.Append(diags...)
}

func (r *metricStreamSettingResource) metricStreamPut(plan metricstream.MetricStreamSettingModel) client.UpsertOwnerMetricsStreamJSONRequestBody {
	return client.UpsertOwnerMetricsStreamJSONRequestBody{
		Provider: common.From(metricstream.ProviderFromPlan(plan.Provider.ValueString())),
		Url:      common.From(plan.URL.ValueString()),
		Token:    common.From(plan.Token.ValueString()),
	}
}

func (r *metricStreamSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	err := common.Delete(func() (*http.Response, error) {
		return r.client.DeleteOwnerMetricsStream(ctx, r.ownerID)
	})
	if err != nil {
		resp.Diagnostics.AddError("unable to delete metric stream settings", err.Error())
		return
	}
}
