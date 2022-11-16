package gobatch

import (
	"time"
)

type JobInstance struct {
	JobInstanceId int64
	JobName       string
	JobKey        string
	JobParams     string
	CreateTime    time.Time
}

type StepContext struct {
	StepContextId int64
	JobInstanceId int64
	StepName      string
	StepContext   *string
	CreateTime    time.Time
	LastUpdated   time.Time
}
