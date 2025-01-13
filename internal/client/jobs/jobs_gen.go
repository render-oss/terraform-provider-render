// Package jobs provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package jobs

import (
	"time"
)

// Defines values for JobStatus.
const (
	Canceled  JobStatus = "canceled"
	Failed    JobStatus = "failed"
	Succeeded JobStatus = "succeeded"
)

// Job defines model for job.
type Job struct {
	CreatedAt    time.Time  `json:"createdAt"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
	Id           JobId      `json:"id"`
	PlanId       string     `json:"planId"`
	ServiceId    string     `json:"serviceId"`
	StartCommand string     `json:"startCommand"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	Status       *JobStatus `json:"status,omitempty"`
}

// JobId defines model for jobId.
type JobId = string

// JobStatus defines model for jobStatus.
type JobStatus string
