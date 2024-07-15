// Package postgres provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.1.0 DO NOT EDIT.
package postgres

import (
	"time"
)

// Defines values for PostgresPlans.
const (
	Custom   PostgresPlans = "custom"
	Free     PostgresPlans = "free"
	Pro      PostgresPlans = "pro"
	ProPlus  PostgresPlans = "pro_plus"
	Standard PostgresPlans = "standard"
	Starter  PostgresPlans = "starter"
)

// Defines values for RecoveryInfoRecoveryStatus.
const (
	AVAILABLE      RecoveryInfoRecoveryStatus = "AVAILABLE"
	BACKUPNOTREADY RecoveryInfoRecoveryStatus = "BACKUP_NOT_READY"
	NOTAVAILABLE   RecoveryInfoRecoveryStatus = "NOT_AVAILABLE"
)

// PostgresBackup defines model for postgresBackup.
type PostgresBackup struct {
	CreatedAt time.Time `json:"createdAt"`
	Id        string    `json:"id"`

	// Url URL to download the Postgres backup
	Url *string `json:"url,omitempty"`
}

// PostgresPlans defines model for postgresPlans.
type PostgresPlans string

// RecoveryInfo defines model for recoveryInfo.
type RecoveryInfo struct {
	// RecoveryStatus Availability of point-in-time recovery.
	RecoveryStatus RecoveryInfoRecoveryStatus `json:"recoveryStatus"`
	StartsAt       *time.Time                 `json:"startsAt,omitempty"`
}

// RecoveryInfoRecoveryStatus Availability of point-in-time recovery.
type RecoveryInfoRecoveryStatus string

// RecoveryInput defines model for recoveryInput.
type RecoveryInput struct {
	// DatadogApiKey Datadog API key to use for monitoring the new database. Defaults to the API key of the original database. Use an empty string to prevent copying of the API key to the new database.
	DatadogApiKey *string `json:"datadogApiKey,omitempty"`

	// EnvironmentId The environment to create the new database in. Defaults to the environment of the original database.
	EnvironmentId *string `json:"environmentId,omitempty"`

	// Plan The plan to use for the new database. Defaults to the same plan as the original database. Cannot be a lower tier plan than the original database.
	Plan *string `json:"plan,omitempty"`

	// RestoreName Name of the new database.
	RestoreName *string `json:"restoreName,omitempty"`

	// RestoreTime The point in time to restore the database to. See `/recovery-info` for restore availability
	RestoreTime time.Time `json:"restoreTime"`
}
