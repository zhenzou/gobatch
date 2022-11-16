package repository

import (
	"context"

	"github.com/chararch/gobatch"
)



type repository struct {
	
}

func (r *repository) CreateJobInstance(jobName string, jobParams map[string]interface{}) (*gobatch.BatchJobInstance, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) SaveJobExecution(execution *gobatch.JobExecution) gobatch.BatchError {
	// TODO implement me
	panic("implement me")
}

func (r *repository) CheckJobStopping(execution *gobatch.JobExecution) (bool, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) SaveStepExecution(ctx context.Context, execution *gobatch.StepExecution) gobatch.BatchError {
	// TODO implement me
	panic("implement me")
}

func (r *repository) UpdateStepStatus(execution *gobatch.StepExecution) gobatch.BatchError {
	// TODO implement me
	panic("implement me")
}

func (r *repository) SaveStepContexts(stepCtx *gobatch.BatchStepContext) gobatch.BatchError {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindJobInstance(jobName string, params map[string]interface{}) (*gobatch.BatchJobInstance, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindLastJobInstanceByName(jobName string) (*gobatch.BatchJobInstance, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindLastJobExecutionByInstance(jobInstance *gobatch.BatchJobInstance) (*gobatch.JobExecution, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindJobExecution(jobExecutionId int64) (*gobatch.JobExecution, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindStepExecutionsByJobExecution(jobExecutionId int64) ([]*gobatch.StepExecution, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindStepExecutionsByName(jobExecutionId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindLastStepExecution(jobInstanceId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindLastCompleteStepExecution(jobInstanceId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

func (r *repository) FindStepContext(jobInstanceId int64, stepName string) (*gobatch.BatchStepContext, gobatch.BatchError) {
	// TODO implement me
	panic("implement me")
}

