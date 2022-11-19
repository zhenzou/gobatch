package gobatch

import (
	"time"
)

type JobInstance struct {
	Id         int64
	JobName    string
	JobKey     string
	JobParams  Parameters
	CreateTime time.Time
}

type StepContext struct {
	Id            int64
	JobInstanceId int64
	StepName      string
	StepContext   *string
	CreateTime    time.Time
	LastUpdated   time.Time
}
