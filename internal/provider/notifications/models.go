package notifications

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client/notifications"
)

type NotificationSettingModel struct {
	EmailEnabled                types.Bool   `tfsdk:"email_enabled"`
	NotificationsToSend         types.String `tfsdk:"notifications_to_send"`
	PreviewNotificationsEnabled types.Bool   `tfsdk:"preview_notifications_enabled"`
	SlackEnabled                types.Bool   `tfsdk:"slack_enabled"`
}

func ModelFromClient(notificationSetting *notifications.NotificationSetting) NotificationSettingModel {
	postgresModel := NotificationSettingModel{
		EmailEnabled:                types.BoolValue(notificationSetting.EmailEnabled),
		NotificationsToSend:         types.StringValue(string(notificationSetting.NotificationsToSend)),
		PreviewNotificationsEnabled: types.BoolValue(notificationSetting.PreviewNotificationsEnabled),
		SlackEnabled:                types.BoolValue(notificationSetting.SlackEnabled),
	}
	return postgresModel
}
