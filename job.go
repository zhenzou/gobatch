package gobatch

import (
	"chararch/gobatch/status"
	"context"
	"github.com/pkg/errors"
	"reflect"
	"runtime/debug"
	"time"
)

type Job interface {
	Name() string
	Start(ctx context.Context, execution *JobExecution) BatchError
	Stop(ctx context.Context, execution *JobExecution) BatchError
	GetSteps() []Step
}

type simpleJob struct {
	name      string
	steps     []Step
	listeners []JobListener
}

func newSimpleJob(name string, steps []Step, listeners []JobListener) *simpleJob {
	return &simpleJob{
		name:      name,
		steps:     steps,
		listeners: listeners,
	}
}

func (job *simpleJob) Name() string {
	return job.name
}

func (job *simpleJob) Start(ctx context.Context, execution *JobExecution) (err BatchError) {
	defer func() {
		if er := recover(); er != nil {
			logger.Error(ctx, "panic in job executing, jobName:%v, jobExecutionId:%v, err:%v, stack:%v", job.name, execution.JobExecutionId, er, debug.Stack())
			execution.JobStatus = status.FAILED
			execution.FailError = errors.Errorf("job execution error:%+v", er)
			execution.EndTime = time.Now()
		}
		if err != nil {
			execution.JobStatus = status.FAILED
			execution.FailError = err
			execution.EndTime = time.Now()
		}
		if err = saveJobExecutions(execution); err != nil {
			logger.Error(ctx, "save job execution failed, jobName:%v, JobExecution:%+v, err:%v", job.name, execution, err)
		}
	}()
	logger.Info(ctx, "start job, jobName:%v, jobExecutionId:%v", job.name, execution.JobExecutionId)
	for _, listener := range job.listeners {
		err = listener.BeforeJob(execution)
		if err != nil {
			logger.Error(ctx, "job listener execute err, jobName:%v, jobExecutionId:%+v, listener:%v, err:%v", job.name, execution.JobExecutionId, reflect.TypeOf(listener).String(), err)
			execution.JobStatus = status.FAILED
			execution.FailError = err
			execution.EndTime = time.Now()
			return nil
		}
	}
	execution.JobStatus = status.STARTED
	execution.StartTime = time.Now()
	if err = saveJobExecutions(execution); err != nil {
		logger.Error(ctx, "save job execution failed, jobName:%v, JobExecution:%+v, err:%v", job.name, execution, err)
		return err
	}
	for _, step := range job.steps {
		e := execStep(ctx, step, execution)
		if e != nil {
			logger.Error(ctx, "execute step failed, jobExecutionId:%v, step:%v, err:%v", execution.JobExecutionId, step.Name(), err)
			if e.Code() == StopError.Code() {
				execution.JobStatus = status.STOPPED
				execution.EndTime = time.Now()
			} else {
				execution.JobStatus = status.FAILED
				execution.EndTime = time.Now()
			}
			break
		}
		if execution.JobStatus == status.FAILED || execution.JobStatus == status.UNKNOWN {
			break
		}
	}
	if execution.JobStatus == "" {
		execution.JobStatus = status.COMPLETED
		execution.EndTime = time.Now()
	}
	for _, listener := range job.listeners {
		err = listener.AfterJob(execution)
		if err != nil {
			logger.Error(ctx, "job listener execute err, jobName:%v, jobExecutionId:%+v, listener:%v, err:%v", job.name, execution.JobExecutionId, reflect.TypeOf(listener).String(), err)
			execution.JobStatus = status.FAILED
			execution.FailError = err
			execution.EndTime = time.Now()
			break
		}
	}
	logger.Info(ctx, "finish job, jobName:%v, jobExecutionId:%v, jobStatus:%v", job.name, execution.JobExecutionId, execution.JobStatus)
	return nil
}

func execStep(ctx context.Context, step Step, execution *JobExecution) (err BatchError) {
	defer func() {
		if err != nil && err.Code() == StopError.Code() {
			logger.Error(ctx, "error in step executing, jobExecutionId:%v, stepName:%v, err:%v", execution.JobExecutionId, step.Name(), err)
		}
	}()
	lastStepExecution, er := findStepExecutionsByName(execution.JobExecutionId, step.Name())
	if er != nil {
		err = er
		logger.Error(ctx, "find last StepExecution failed, jobExecutionId:%v, stepName:%v, err:%v", execution.JobExecutionId, step.Name(), er)
		return er
	}
	if lastStepExecution != nil && lastStepExecution.StepStatus == status.COMPLETED {
		logger.Info(ctx, "skip completed step, jobExecutionId:%v, stepName:%v", execution.JobExecutionId, step.Name())
		return nil
	}
	if lastStepExecution != nil && (lastStepExecution.StepStatus == status.STARTING || lastStepExecution.StepStatus == status.STARTED || lastStepExecution.StepStatus == status.STOPPING) {
		logger.Error(ctx, "last StepExecution is in progress, jobExecutionId:%v, stepName:%v", execution.JobExecutionId, step.Name())
		return ConcurrentError
	}
	stepExecution := &StepExecution{
		StepName:             step.Name(),
		StepStatus:           status.STARTING,
		StepContext:          NewBatchContext(),
		StepExecutionContext: NewBatchContext(),
		JobExecution:         execution,
		CreateTime:           time.Now(),
	}
	if lastStepExecution != nil {
		stepExecution.StepContext.Merge(lastStepExecution.StepContext)
		stepExecution.StepExecutionContext.Merge(lastStepExecution.StepExecutionContext)
	}
	e := saveStepExecution(ctx, stepExecution)
	if e != nil {
		logger.Error(ctx, "save step execution failed, jobExecutionId:%v, stepName:%v, err:%v", execution.JobExecutionId, step.Name(), e)
		err = e
		return err
	}
	execution.AddStepExecution(stepExecution)
	err = step.Exec(ctx, stepExecution)
	if err != nil || stepExecution.StepStatus != status.COMPLETED {
		logger.Error(ctx, "step executing failed, jobExecutionId:%v, stepName:%v, stepStatus:%v, err:%v", execution.JobExecutionId, step.Name(), stepExecution.StepStatus, e)
		execution.JobStatus = stepExecution.StepStatus
		if err != nil && stepExecution.StepStatus != status.FAILED {
			stepExecution.StepStatus = status.FAILED
			stepExecution.FailError = err
			stepExecution.EndTime = time.Now()
			execution.JobStatus = status.FAILED
		}
		execution.FailError = err
		execution.EndTime = time.Now()
		e = saveStepExecution(ctx, stepExecution)
		if e != nil {
			logger.Error(ctx, "save step execution failed, jobExecutionId:%v, stepName:%v, err:%v", execution.JobExecutionId, step.Name(), e)
			err = e
			return err
		}
	} else {
		stepExecution.StepStatus = status.COMPLETED
		stepExecution.EndTime = time.Now()
		e = saveStepExecution(ctx, stepExecution)
		if e != nil {
			logger.Error(ctx, "save step execution failed, jobExecutionId:%v, stepName:%v, err:%v", execution.JobExecutionId, step.Name(), e)
			err = e
			return err
		}
	}
	return nil
}

func (job *simpleJob) Stop(ctx context.Context, execution *JobExecution) BatchError {
	logger.Info(ctx, "stop job, jobName:%v, jobExecutionId:%v, jobStatus:%v", job.name, execution.JobExecutionId, execution.JobStatus)
	execution.JobStatus = status.STOPPING
	return saveJobExecutions(execution)
}

func (job *simpleJob) GetSteps() []Step {
	return job.steps
}
