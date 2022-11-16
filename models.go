package gobatch

import (
	"time"
)

type BatchJobInstance struct {
	JobInstanceId int64
	JobName       string
	JobKey        string
	JobParams     string
	CreateTime    time.Time
}

// the following models is db model, will replace by entity

type BatchJobExecution struct {
	JobExecutionId int64
	JobInstanceId  int64
	JobName        string
	CreateTime     time.Time
	StartTime      time.Time
	EndTime        time.Time
	Status         string
	ExitCode       string
	ExitMessage    *string
	LastUpdated    time.Time
	Version        int64
}

type BatchStepExecution struct {
	StepExecutionId  int64
	JobExecutionId   int64
	JobInstanceId    int64
	JobName          string
	StepName         string
	CreateTime       time.Time
	StartTime        time.Time
	EndTime          time.Time
	Status           string
	ReadCount        int64
	WriteCount       int64
	CommitCount      int64
	FilterCount      int64
	ReadSkipCount    int64
	WriteSkipCount   int64
	ProcessSkipCount int64
	RollbackCount    int64
	ExecutionContext string
	StepContextId    int64
	ExitCode         string
	ExitMessage      *string
	LastUpdated      time.Time
	Version          int64
}

type BatchStepContext struct {
	StepContextId int64
	JobInstanceId int64
	StepName      string
	StepContext   *string
	CreateTime    time.Time
	LastUpdated   time.Time
}
