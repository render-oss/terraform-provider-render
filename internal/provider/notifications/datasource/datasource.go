package datasource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/notifications"

	"terraform-provider-render/internal/client"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &notificationSettingDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationSettingDataSource{}
)

// NewNotificationSettingDataSource is a helper function to simplify the provider implementation.
func NewNotificationSettingDataSource() datasource.DataSource {
	return &notificationSettingDataSource{}
}

// notificationSettingDataSource is the data source implementation.common.
type notificationSettingDataSource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured client to the data source.
func (d *notificationSettingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data := rendertypes.ConfigureDatasource(req, resp)
	if data == nil {
		return
	}

	d.client = data.Client
	d.ownerID = data.OwnerID
}

// Metadata returns the data source type name.
func (d *notificationSettingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_setting"
}

// Schema defines the schema for the data source.
func (d *notificationSettingDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = NotificationSettingDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *notificationSettingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan notifications.NotificationSettingModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var notifs client.NotificationSetting
	err := common.Get(func() (*http.Response, error) {
		return d.client.GetOwnerNotificationSettings(ctx, d.ownerID)
	}, &notifs)
	if err != nil {
		resp.Diagnostics.AddError("unable to get notificationSetting", err.Error())
		return
	}

	resp.State.Set(ctx, notifications.ModelFromClient(&notifs))
}
