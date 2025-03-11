// Package eventtypes provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package eventtypes

// Defines values for EventType.
const (
	EventTypeAutoscalingConfigChanged     EventType = "autoscaling_config_changed"
	EventTypeAutoscalingEnded             EventType = "autoscaling_ended"
	EventTypeAutoscalingStarted           EventType = "autoscaling_started"
	EventTypeBranchDeleted                EventType = "branch_deleted"
	EventTypeBuildEnded                   EventType = "build_ended"
	EventTypeBuildPlanChanged             EventType = "build_plan_changed"
	EventTypeBuildStarted                 EventType = "build_started"
	EventTypeCommitIgnored                EventType = "commit_ignored"
	EventTypeCronJobRunEnded              EventType = "cron_job_run_ended"
	EventTypeCronJobRunStarted            EventType = "cron_job_run_started"
	EventTypeDeployEnded                  EventType = "deploy_ended"
	EventTypeDeployStarted                EventType = "deploy_started"
	EventTypeDiskCreated                  EventType = "disk_created"
	EventTypeDiskDeleted                  EventType = "disk_deleted"
	EventTypeDiskUpdated                  EventType = "disk_updated"
	EventTypeImagePullFailed              EventType = "image_pull_failed"
	EventTypeInstanceCountChanged         EventType = "instance_count_changed"
	EventTypeJobRunEnded                  EventType = "job_run_ended"
	EventTypeKeyValueAvailable            EventType = "key_value_available"
	EventTypeKeyValueConfigRestart        EventType = "key_value_config_restart"
	EventTypeKeyValueUnhealthy            EventType = "key_value_unhealthy"
	EventTypeMaintenanceEnded             EventType = "maintenance_ended"
	EventTypeMaintenanceModeEnabled       EventType = "maintenance_mode_enabled"
	EventTypeMaintenanceModeUriUpdated    EventType = "maintenance_mode_uri_updated"
	EventTypeMaintenanceStarted           EventType = "maintenance_started"
	EventTypePlanChanged                  EventType = "plan_changed"
	EventTypePostgresAvailable            EventType = "postgres_available"
	EventTypePostgresBackupCompleted      EventType = "postgres_backup_completed"
	EventTypePostgresBackupStarted        EventType = "postgres_backup_started"
	EventTypePostgresClusterLeaderChanged EventType = "postgres_cluster_leader_changed"
	EventTypePostgresCreated              EventType = "postgres_created"
	EventTypePostgresDiskSizeChanged      EventType = "postgres_disk_size_changed"
	EventTypePostgresHaStatusChanged      EventType = "postgres_ha_status_changed"
	EventTypePostgresReadReplicasChanged  EventType = "postgres_read_replicas_changed"
	EventTypePostgresRestarted            EventType = "postgres_restarted"
	EventTypePostgresUnavailable          EventType = "postgres_unavailable"
	EventTypePostgresUpgradeFailed        EventType = "postgres_upgrade_failed"
	EventTypePostgresUpgradeStarted       EventType = "postgres_upgrade_started"
	EventTypePostgresUpgradeSucceeded     EventType = "postgres_upgrade_succeeded"
	EventTypePreDeployEnded               EventType = "pre_deploy_ended"
	EventTypePreDeployStarted             EventType = "pre_deploy_started"
	EventTypeServerAvailable              EventType = "server_available"
	EventTypeServerFailed                 EventType = "server_failed"
	EventTypeServerHardwareFailure        EventType = "server_hardware_failure"
	EventTypeServerRestarted              EventType = "server_restarted"
	EventTypeServerUnhealthy              EventType = "server_unhealthy"
	EventTypeServiceResumed               EventType = "service_resumed"
	EventTypeServiceSuspended             EventType = "service_suspended"
	EventTypeZeroDowntimeRedeployEnded    EventType = "zero_downtime_redeploy_ended"
	EventTypeZeroDowntimeRedeployStarted  EventType = "zero_downtime_redeploy_started"
)

// Defines values for ServiceEventType.
const (
	ServiceEventTypeAutoscalingConfigChanged    ServiceEventType = "autoscaling_config_changed"
	ServiceEventTypeAutoscalingEnded            ServiceEventType = "autoscaling_ended"
	ServiceEventTypeAutoscalingStarted          ServiceEventType = "autoscaling_started"
	ServiceEventTypeBranchDeleted               ServiceEventType = "branch_deleted"
	ServiceEventTypeBuildEnded                  ServiceEventType = "build_ended"
	ServiceEventTypeBuildPlanChanged            ServiceEventType = "build_plan_changed"
	ServiceEventTypeBuildStarted                ServiceEventType = "build_started"
	ServiceEventTypeCommitIgnored               ServiceEventType = "commit_ignored"
	ServiceEventTypeCronJobRunEnded             ServiceEventType = "cron_job_run_ended"
	ServiceEventTypeCronJobRunStarted           ServiceEventType = "cron_job_run_started"
	ServiceEventTypeDeployEnded                 ServiceEventType = "deploy_ended"
	ServiceEventTypeDeployStarted               ServiceEventType = "deploy_started"
	ServiceEventTypeDiskCreated                 ServiceEventType = "disk_created"
	ServiceEventTypeDiskDeleted                 ServiceEventType = "disk_deleted"
	ServiceEventTypeDiskUpdated                 ServiceEventType = "disk_updated"
	ServiceEventTypeImagePullFailed             ServiceEventType = "image_pull_failed"
	ServiceEventTypeInitialDeployHookEnded      ServiceEventType = "initial_deploy_hook_ended"
	ServiceEventTypeInitialDeployHookStarted    ServiceEventType = "initial_deploy_hook_started"
	ServiceEventTypeInstanceCountChanged        ServiceEventType = "instance_count_changed"
	ServiceEventTypeJobRunEnded                 ServiceEventType = "job_run_ended"
	ServiceEventTypeMaintenanceEnded            ServiceEventType = "maintenance_ended"
	ServiceEventTypeMaintenanceModeEnabled      ServiceEventType = "maintenance_mode_enabled"
	ServiceEventTypeMaintenanceModeUriUpdated   ServiceEventType = "maintenance_mode_uri_updated"
	ServiceEventTypeMaintenanceStarted          ServiceEventType = "maintenance_started"
	ServiceEventTypePlanChanged                 ServiceEventType = "plan_changed"
	ServiceEventTypePreDeployEnded              ServiceEventType = "pre_deploy_ended"
	ServiceEventTypePreDeployStarted            ServiceEventType = "pre_deploy_started"
	ServiceEventTypeServerAvailable             ServiceEventType = "server_available"
	ServiceEventTypeServerFailed                ServiceEventType = "server_failed"
	ServiceEventTypeServerHardwareFailure       ServiceEventType = "server_hardware_failure"
	ServiceEventTypeServerRestarted             ServiceEventType = "server_restarted"
	ServiceEventTypeServerUnhealthy             ServiceEventType = "server_unhealthy"
	ServiceEventTypeServiceResumed              ServiceEventType = "service_resumed"
	ServiceEventTypeServiceSuspended            ServiceEventType = "service_suspended"
	ServiceEventTypeSuspenderAdded              ServiceEventType = "suspender_added"
	ServiceEventTypeSuspenderRemoved            ServiceEventType = "suspender_removed"
	ServiceEventTypeZeroDowntimeRedeployEnded   ServiceEventType = "zero_downtime_redeploy_ended"
	ServiceEventTypeZeroDowntimeRedeployStarted ServiceEventType = "zero_downtime_redeploy_started"
)

// EventType defines model for eventType.
type EventType string

// ServiceEventType defines model for serviceEventType.
type ServiceEventType string
