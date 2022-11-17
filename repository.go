package gobatch

import (
	"context"
)

type Repository interface {
	CreateJobInstance(jobName string, jobParams map[string]interface{}) (*JobInstance, BatchError)
	SaveJobExecution(execution *JobExecution) BatchError
	CheckJobStopping(execution *JobExecution) (bool, BatchError)

	SaveStepExecution(ctx context.Context, execution *StepExecution) BatchError
	UpdateStepStatus(execution *StepExecution) BatchError
	SaveStepContexts(stepCtx *StepContext) BatchError

	FindJobInstance(jobName string, params map[string]interface{}) (*JobInstance, BatchError)
	FindLastJobInstanceByName(jobName string) (*JobInstance, BatchError)
	FindLastJobExecutionByInstance(jobInstance *JobInstance) (*JobExecution, BatchError)
	FindJobExecution(jobExecutionId int64) (*JobExecution, BatchError)
	FindStepExecutionsByJobExecution(jobExecutionId int64) ([]*StepExecution, BatchError)
	FindStepExecutionsByName(jobExecutionId int64, stepName string) (*StepExecution, BatchError)
	FindLastStepExecution(jobInstanceId int64, stepName string) (*StepExecution, BatchError)
	FindLastCompleteStepExecution(jobInstanceId int64, stepName string) (*StepExecution, BatchError)
	FindStepContext(jobInstanceId int64, stepName string) (*StepContext, BatchError)
}
