package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/project"
	rendertypes "terraform-provider-render/internal/provider/types"
)

var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// The data source exposes a subset of the resource's environment attributes.
type projectDataSourceModel struct {
	Id           types.String                          `tfsdk:"id"`
	Name         types.String                          `tfsdk:"name"`
	Environments map[string]environmentDataSourceModel `tfsdk:"environments"`
}

type environmentDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ProtectedStatus types.String `tfsdk:"protected_status"`
	NetworkIsolated types.Bool   `tfsdk:"network_isolated"`
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ProjectDataSourceSchema(ctx)
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var proj projectDataSourceModel
	diags := req.Config.Get(ctx, &proj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectModel, err := project.Read(ctx, d.client, project.ProjectModel{Id: proj.Id})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project", "Could not read project, unexpected error: "+err.Error(),
		)
		return
	}

	proj.Name = projectModel.Name
	proj.Environments = make(map[string]environmentDataSourceModel, len(projectModel.Environments))
	for key, env := range projectModel.Environments {
		proj.Environments[key] = environmentDataSourceModel{
			Id:              env.Id,
			Name:            env.Name,
			ProtectedStatus: env.ProtectedStatus,
			NetworkIsolated: env.NetworkIsolated,
		}
	}
	diags = resp.State.Set(ctx, &proj)
	resp.Diagnostics.Append(diags...)
}
