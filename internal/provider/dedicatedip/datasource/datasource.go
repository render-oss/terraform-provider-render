package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/dedicatedip"
	rendertypes "terraform-provider-render/internal/provider/types"
)

var (
	_ datasource.DataSource              = &dedicatedIPDataSource{}
	_ datasource.DataSourceWithConfigure = &dedicatedIPDataSource{}
)

func NewDedicatedIPDataSource() datasource.DataSource {
	return &dedicatedIPDataSource{}
}

type dedicatedIPDataSource struct {
	client *client.ClientWithResponses
}

func (d *dedicatedIPDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}
	d.client = data.Client
}

func (d *dedicatedIPDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dedicated_ip"
}

func (d *dedicatedIPDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema(ctx)
}

func (d *dedicatedIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg dedicatedip.Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var fetched client.DedicatedIP
	if err := common.Get(func() (*http.Response, error) {
		return d.client.RetrieveDedicatedIp(ctx, cfg.ID.ValueString())
	}, &fetched); err != nil {
		resp.Diagnostics.AddError("Unable to get dedicated IP", err.Error())
		return
	}

	state := dedicatedip.ModelFromClient(&fetched, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
