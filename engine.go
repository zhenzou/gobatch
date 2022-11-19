package gobatch

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/chararch/gobatch/util"
)

type Engine interface {
	Register(job Job) error
	Unregister(job Job)
	Start(ctx context.Context, jobName string, params string) (int64, error)
	StartAsync(ctx context.Context, jobName string, params string) (int64, error)
	Stop(ctx context.Context, jobId interface{}) error
	Restart(ctx context.Context, jobId interface{}) (int64, error)
	RestartAsync(ctx context.Context, jobId interface{}) (int64, error)
}

func NewEngine(repository Repository) Engine {
	return &engine{
		jobRegistry: map[string]Job{},
		repository:  repository,
	}
}

type engine struct {
	repository  Repository
	jobRegistry map[string]Job
}

// Register register job to gobatch
func (e *engine) Register(job Job) error {
	if _, ok := e.jobRegistry[job.Name()]; ok {
		return fmt.Errorf("job with name:%v has already been registered", job.Name())
	}
	e.jobRegistry[job.Name()] = job
	return nil
}

// Unregister unregister job to gobatch
func (e *engine) Unregister(job Job) {
	delete(e.jobRegistry, job.Name())
}

// Start start job by job name and params
func (e *engine) Start(ctx context.Context, jobName string, params string) (int64, error) {
	return e.doStart(ctx, jobName, params, false)
}

// StartAsync start job by job name and params asynchronously
func (e *engine) StartAsync(ctx context.Context, jobName string, params string) (int64, error) {
	return e.doStart(ctx, jobName, params, true)
}

func (e *engine) doStart(ctx context.Context, jobName string, params string, async bool) (int64, error) {
	if job, ok := e.jobRegistry[jobName]; ok {
		jobParams, err := ParseJobParams(params)
		if err != nil {
			DefaultLogger.Error(ctx, "parse job params error, jobName:%v, params:%v, err:%v", jobName, params, err)
			return -1, err
		}
		jobInstance, err := e.repository.FindJobInstance(jobName, jobParams)
		if err != nil {
			DefaultLogger.Error(ctx, "find JobInstance error, jobName:%v, params:%v, err:%v", jobName, params, err)
			return -1, err
		}
		if jobInstance == nil {
			jobInstance, err = e.repository.CreateJobInstance(jobName, jobParams)
			if err != nil {
				DefaultLogger.Error(ctx, "find JobInstance error, jobName:%v, params:%v, err:%v", jobName, params, err)
				return -1, err
			}
		}
		jobExecution, err := e.repository.FindLastJobExecutionByInstance(jobInstance)
		if err != nil {
			DefaultLogger.Error(ctx, "find last JobExecution error, jobName:%v, jobInstanceId:%v, err:%v", jobName, jobInstance.JobInstanceId, err)
			return -1, err
		}
		if jobExecution != nil {
			lastExecution := jobExecution
			jobStatus := lastExecution.JobStatus
			if jobStatus == STARTING || jobStatus == STARTED || jobStatus == STOPPING || jobStatus == UNKNOWN {
				DefaultLogger.Error(ctx, "the job is in executing or exit from last execution abnormally, can not restart, jobName:%v, status:%v", jobName, jobStatus)
				return -1, errors.Errorf("the job is in executing or exit from last execution abnormally, can not restart, jobName:%v, status:%v", jobName, jobStatus)
			}
			// find step executions & check step execution status
			stepExecutions, err := e.repository.FindStepExecutionsByJobExecution(lastExecution.JobExecutionId)
			if err != nil {
				DefaultLogger.Error(ctx, "find last StepExecution error, jobName:%v, jobExecutionId:%v, err:%v", jobName, lastExecution.JobExecutionId, err)
				return -1, err
			}
			for _, stepExecution := range stepExecutions {
				if stepExecution.StepStatus == UNKNOWN {
					DefaultLogger.Error(ctx, "can not restart a job that has step with unknown status, job:%v step:%v", jobName, stepExecution.StepName)
					return -1, errors.Errorf("can not restart a job that has step with unknown status, job:%v step:%v", jobName, stepExecution.StepName)
				}
			}
		}
		// new
		execution := &JobExecution{
			JobInstanceId:  jobInstance.JobInstanceId,
			JobName:        jobName,
			JobParams:      jobParams,
			JobStatus:      STARTING,
			StepExecutions: make([]*StepExecution, 0),
			JobContext:     NewBatchContext(),
			CreateTime:     time.Now(),
		}
		err = e.repository.SaveJobExecution(execution)
		if err != nil {
			DefaultLogger.Error(ctx, "save job execution failed, jobName:%v, JobExecution:%+v, err:%v", jobName, execution, err)
			return -1, err
		}
		future := jobPool.Submit(ctx, func() (interface{}, error) {
			er := job.Start(ctx, execution)
			return nil, er
		})
		DefaultLogger.Info(ctx, "job started, jobName:%v, jobExecutionId:%v", jobName, execution.JobExecutionId)
		if async {
			return execution.JobExecutionId, nil
		} else {
			if _, er := future.Get(); er != nil {
				return execution.JobExecutionId, er
			} else {
				return execution.JobExecutionId, nil
			}
		}
	} else {
		DefaultLogger.Error(ctx, "can not find job with name:%v", jobName)
		return -1, errors.Errorf("can not find job with name:%v", jobName)
	}
}

// Stop stop job by job name or job execution id
func (e *engine) Stop(ctx context.Context, jobId interface{}) error {
	switch id := jobId.(type) {
	case string:
		if job, ok := e.jobRegistry[id]; ok {
			// find executions by jobName, then stop
			jobInstance, err := e.repository.FindLastJobInstanceByName(job.Name())
			if err != nil {
				DefaultLogger.Error(ctx, "find last JobInstance error, jobName:%v, err:%v", job.Name(), err)
				return err
			}
			if jobInstance != nil {
				execution, err := e.repository.FindLastJobExecutionByInstance(jobInstance)
				if err != nil {
					DefaultLogger.Error(ctx, "find last JobExecution error, jobName:%v, jobInstanceId:%v, err:%v", job.Name(), jobInstance.JobInstanceId, err)
					return err
				}
				if execution != nil && execution.JobStatus == STARTING || execution.JobStatus == STARTED {
					DefaultLogger.Info(ctx, "job will be stopped, jobName:%v, jobExecutionId:%v", job.Name(), execution.JobExecutionId)
					return job.Stop(ctx, execution)
				} else {
					DefaultLogger.Error(ctx, "there is no running job instance with name:%v to stop", id)
					return errors.Errorf("there is no running job instance with name:%v to stop", id)
				}
			} else {
				DefaultLogger.Error(ctx, "there is no running job instance with name:%v to stop", id)
				return errors.Errorf("there is no running job instance with name:%v to stop", id)
			}
		} else {
			DefaultLogger.Error(ctx, "can not find job with name:%v", id)
			return errors.Errorf("can not find job with name:%v", id)
		}
	case int64:
		// find executions by execution id, if found then stop
		execution, err := e.repository.FindJobExecution(id)
		if err != nil {
			DefaultLogger.Error(ctx, "find JobExecution by jobExecutionId error, jobExecutionId:%v, err:%v", id, err)
			return err
		}
		if execution == nil {
			DefaultLogger.Error(ctx, "can not find job execution with execution id:%v", id)
			return errors.Errorf("can not find job execution with execution id:%v", id)
		}
		if job, ok := e.jobRegistry[execution.JobName]; ok {
			return job.Stop(ctx, execution)
		} else {
			DefaultLogger.Error(ctx, "can not find job with name:%v", execution.JobName)
			return errors.Errorf("can not find job with name:%v", execution.JobName)
		}
	}
	DefaultLogger.Error(ctx, "job identifier:%v is either job name or job execution id", jobId)
	return errors.Errorf("job identifier:%v is either job name or job execution id", jobId)
}

// Restart restart job by job name or job execution id
func (e *engine) Restart(ctx context.Context, jobId interface{}) (int64, error) {
	return e.doRestart(ctx, jobId, false)
}

// RestartAsync restart job by job name or job execution id asynchronously
func (e *engine) RestartAsync(ctx context.Context, jobId interface{}) (int64, error) {
	return e.doRestart(ctx, jobId, true)
}

func (e *engine) doRestart(ctx context.Context, jobId interface{}, async bool) (int64, error) {
	// find executions, ensure no running instance and then start
	switch id := jobId.(type) {
	case string:
		if job, ok := e.jobRegistry[id]; ok {
			// find executions by jobName, if count==1 then stop
			jobInstance, err := e.repository.FindLastJobInstanceByName(job.Name())
			if err != nil {
				DefaultLogger.Error(ctx, "find last JobInstance error, jobName:%v, err:%v", job.Name(), err)
				return -1, err
			}
			if jobInstance != nil {
				return e.doStart(ctx, job.Name(), jobInstance.JobParams, async)
			} else {
				return e.doStart(ctx, job.Name(), "", async)
			}
		}
		DefaultLogger.Error(ctx, "can not find job with name:%v", id)
		return -1, errors.Errorf("can not find job with name:%v", id)
	case int64:
		// find executions by execution id, then start
		execution, err := e.repository.FindJobExecution(id)
		if err != nil {
			DefaultLogger.Error(ctx, "find JobExecution by jobExecutionId error, jobExecutionId:%v, err:%v", id, err)
			return -1, err
		}
		if execution == nil {
			DefaultLogger.Error(ctx, "can not find job execution with execution id:%v", id)
			return -1, errors.Errorf("can not find job execution with execution id:%v", id)
		}
		if job, ok := e.jobRegistry[execution.JobName]; ok {
			params, _ := util.JsonString(execution.JobParams)
			return e.doStart(ctx, job.Name(), params, async)
		}
		DefaultLogger.Error(ctx, "can not find job with name:%v", execution.JobName)
		return -1, errors.Errorf("can not find job with name:%v", execution.JobName)
	}
	DefaultLogger.Error(ctx, "job identifier:%v is either job name or job execution id", jobId)
	return -1, errors.Errorf("job identifier:%v is either job name or job execution id", jobId)
}

func ParseJobParams(params string) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	if len(params) == 0 {
		return ret, nil
	}
	err := util.ParseJson(params, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
