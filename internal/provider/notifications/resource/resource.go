package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/notifications"
	rendertypes "terraform-provider-render/internal/provider/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &notificationSettingResource{}
	_ resource.ResourceWithConfigure = &notificationSettingResource{}
)

// NewNotificationSettingResource is a helper function to simplify the provider implementation.
func NewNotificationSettingResource() resource.Resource {
	return &notificationSettingResource{}
}

// notificationSettingResource is the resource implementation.
type notificationSettingResource struct {
	client  *client.ClientWithResponses
	ownerID string
}

// Configure adds the provider configured Client to the resource.
func (r *notificationSettingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := rendertypes.ConfigureResource(req, resp)
	if data == nil {
		return
	}

	r.client = data.Client
	r.ownerID = data.OwnerID
}

// Metadata returns the resource type name.
func (r *notificationSettingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_setting"
}

// Schema defines the schema for the resource.
func (r *notificationSettingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = NotificationSettingResourceSchema(ctx)
}

// Create a new resource.
func (r *notificationSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notifications.NotificationSettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var notifs client.NotificationSetting

	patch := r.notificationPatch(plan)

	err := common.Update(func() (*http.Response, error) {
		return r.client.PatchOwnerNotificationSettings(ctx, r.ownerID, patch)
	}, &notifs)
	if err != nil {
		resp.Diagnostics.AddError("unable to create notification settings", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, notifications.ModelFromClient(&notifs))
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *notificationSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var settings notifications.NotificationSettingModel
	diags := req.State.Get(ctx, &settings)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	var notifs client.NotificationSetting

	err := common.Get(func() (*http.Response, error) {
		return r.client.GetOwnerNotificationSettings(ctx, r.ownerID)
	}, &notifs)
	if common.IsNotFoundErr(err) {
		common.EmitNotFoundWarning(r.ownerID, &diags)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("unable to get notification settings", err.Error())
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, notifications.ModelFromClient(&notifs))
	resp.Diagnostics.Append(diags...)
}

func (r *notificationSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notifications.NotificationSettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var notifs client.NotificationSetting

	patch := r.notificationPatch(plan)

	err := common.Update(func() (*http.Response, error) {
		return r.client.PatchOwnerNotificationSettings(ctx, r.ownerID, patch)
	}, &notifs)
	if err != nil {
		resp.Diagnostics.AddError("unable to update notification settings", err.Error())
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, notifications.ModelFromClient(&notifs))
	resp.Diagnostics.Append(diags...)
}

func (r *notificationSettingResource) notificationPatch(plan notifications.NotificationSettingModel) client.NotificationSettingPATCH {
	emailEnabled := plan.EmailEnabled.ValueBoolPointer()
	if plan.EmailEnabled.IsUnknown() {
		emailEnabled = nil
	}

	prevNotifsEnabled := plan.PreviewNotificationsEnabled.ValueBoolPointer()
	if plan.PreviewNotificationsEnabled.IsUnknown() {
		prevNotifsEnabled = nil
	}

	var notifsToSend *client.NotifySettingV2
	if plan.NotificationsToSend.IsUnknown() || plan.NotificationsToSend.IsNull() {
		notifsToSend = nil
	} else {
		notifsToSend = common.From(client.NotifySettingV2(plan.NotificationsToSend.ValueString()))
	}
	return client.NotificationSettingPATCH{
		EmailEnabled:                emailEnabled,
		NotificationsToSend:         notifsToSend,
		PreviewNotificationsEnabled: prevNotifsEnabled,
	}
}

func (r *notificationSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// This resource always exists, so nothing to do here. We just want
	// to disassociate the resource from the state.
}
