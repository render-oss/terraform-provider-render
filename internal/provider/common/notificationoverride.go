package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
)

var notificationTypes = map[string]attr.Type{
	"preview_notifications_enabled": types.StringType,
	"notifications_to_send":         types.StringType,
}

func NotificationOverrideFromClient(client *client.NotificationOverride, diags diag.Diagnostics) types.Object {
	if client == nil {
		return types.ObjectNull(notificationTypes)
	}

	objectValue, objectDiags := types.ObjectValue(
		notificationTypes,
		map[string]attr.Value{
			"preview_notifications_enabled": types.StringValue(string(client.PreviewNotificationsEnabled)),
			"notifications_to_send":         types.StringValue(string(client.NotificationsToSend)),
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func NotificationOverrideToClient(model types.Object) (*client.NotificationServiceOverridePATCH, error) {
	if model.IsNull() {
		return nil, nil
	}

	attrs := model.Attributes()

	var previewNotificationEnabled *client.NotifyPreviewOverride
	if attrs["preview_notifications_enabled"] != nil && !attrs["preview_notifications_enabled"].IsNull() && !attrs["preview_notifications_enabled"].IsUnknown() {
		str, ok := attrs["preview_notifications_enabled"].(types.String)
		if !ok {
			// This should never happen
			return nil, fmt.Errorf("unexpected type for preview_notifications_enabled: %T", attrs["preview_notifications_enabled"])
		}
		previewNotificationEnabled = From(client.NotifyPreviewOverride(str.ValueString()))
	}
	var notificationsToSend *client.NotifyOverride
	if attrs["notifications_to_send"] != nil && !attrs["notifications_to_send"].IsNull() && !attrs["notifications_to_send"].IsUnknown() {
		str, ok := attrs["notifications_to_send"].(types.String)
		if !ok {
			// This should never happen
			return nil, fmt.Errorf("unexpected type for notifications_to_send: %T", attrs["notifications_to_send"])
		}
		notificationsToSend = From(client.NotifyOverride(str.ValueString()))
	}

	return &client.NotificationServiceOverridePATCH{
		PreviewNotificationsEnabled: previewNotificationEnabled,
		NotificationsToSend:         notificationsToSend,
	}, nil
}
