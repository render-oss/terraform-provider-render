// Package eventtypes provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package eventtypes

// Defines values for EventType.
const (
	AutoscalingConfigChanged    EventType = "autoscaling_config_changed"
	AutoscalingEnded            EventType = "autoscaling_ended"
	AutoscalingStarted          EventType = "autoscaling_started"
	BranchDeleted               EventType = "branch_deleted"
	BuildEnded                  EventType = "build_ended"
	BuildPlanChanged            EventType = "build_plan_changed"
	BuildStarted                EventType = "build_started"
	CommitIgnored               EventType = "commit_ignored"
	CronJobRunEnded             EventType = "cron_job_run_ended"
	CronJobRunStarted           EventType = "cron_job_run_started"
	DeployEnded                 EventType = "deploy_ended"
	DeployStarted               EventType = "deploy_started"
	DiskCreated                 EventType = "disk_created"
	DiskDeleted                 EventType = "disk_deleted"
	DiskUpdated                 EventType = "disk_updated"
	ImagePullFailed             EventType = "image_pull_failed"
	InitialDeployHookEnded      EventType = "initial_deploy_hook_ended"
	InitialDeployHookStarted    EventType = "initial_deploy_hook_started"
	InstanceCountChanged        EventType = "instance_count_changed"
	JobRunEnded                 EventType = "job_run_ended"
	MaintenanceEnded            EventType = "maintenance_ended"
	MaintenanceModeEnabled      EventType = "maintenance_mode_enabled"
	MaintenanceModeUriUpdated   EventType = "maintenance_mode_uri_updated"
	MaintenanceStarted          EventType = "maintenance_started"
	PlanChanged                 EventType = "plan_changed"
	PreDeployEnded              EventType = "pre_deploy_ended"
	PreDeployStarted            EventType = "pre_deploy_started"
	ServerAvailable             EventType = "server_available"
	ServerFailed                EventType = "server_failed"
	ServerHardwareFailure       EventType = "server_hardware_failure"
	ServerRestarted             EventType = "server_restarted"
	ServerUnhealthy             EventType = "server_unhealthy"
	ServiceResumed              EventType = "service_resumed"
	ServiceSuspended            EventType = "service_suspended"
	SuspenderAdded              EventType = "suspender_added"
	SuspenderRemoved            EventType = "suspender_removed"
	ZeroDowntimeRedeployEnded   EventType = "zero_downtime_redeploy_ended"
	ZeroDowntimeRedeployStarted EventType = "zero_downtime_redeploy_started"
)

// EventType defines model for eventType.
type EventType string
