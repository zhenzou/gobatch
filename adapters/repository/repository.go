package gobatch

import (
	"context"

	"github.com/chararch/gobatch"
)

type Repository interface {
	CreateJobInstance(jobName string, jobParams map[string]interface{}) (*gobatch.BatchJobInstance, gobatch.BatchError)
	SaveJobExecution(execution *gobatch.JobExecution) gobatch.BatchError
	CheckJobStopping(execution *gobatch.JobExecution) (bool, gobatch.BatchError)

	SaveStepExecution(ctx context.Context, execution *gobatch.StepExecution) gobatch.BatchError
	UpdateStepStatus(execution *gobatch.StepExecution) gobatch.BatchError
	SaveStepContexts(stepCtx *gobatch.BatchStepContext) gobatch.BatchError

	FindJobInstance(jobName string, params map[string]interface{}) (*gobatch.BatchJobInstance, gobatch.BatchError)
	FindLastJobInstanceByName(jobName string) (*gobatch.BatchJobInstance, gobatch.BatchError)
	FindLastJobExecutionByInstance(jobInstance *gobatch.BatchJobInstance) (*gobatch.JobExecution, gobatch.BatchError)
	FindJobExecution(jobExecutionId int64) (*gobatch.JobExecution, gobatch.BatchError)
	FindStepExecutionsByJobExecution(jobExecutionId int64) ([]*gobatch.StepExecution, gobatch.BatchError)
	FindStepExecutionsByName(jobExecutionId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError)
	FindLastStepExecution(jobInstanceId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError)
	FindLastCompleteStepExecution(jobInstanceId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError)
	FindStepContext(jobInstanceId int64, stepName string) (*gobatch.BatchStepContext, gobatch.BatchError)
}
