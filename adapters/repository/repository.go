// Package repository the default repository implementation
package repository

import (
	"bytes"
	"context"
	"database/sql"
	"time"

	"github.com/chararch/gobatch"
	"github.com/chararch/gobatch/util"
)

func New(db *sql.DB, logger gobatch.Logger) gobatch.Repository {
	return &repositoryImpl{
		db:     db,
		logger: logger,
	}
}

type repositoryImpl struct {
	db     *sql.DB
	logger gobatch.Logger
}

// FindJobInstance load or save job instance by name & parameters
func (r *repositoryImpl) FindJobInstance(jobName string, params map[string]interface{}) (*gobatch.JobInstance, gobatch.BatchError) {
	str, err := util.JsonString(params)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeGeneral, "jsonify job params failed", err)
	}
	key := util.MD5(str)
	rows, err := r.db.Query("SELECT JOB_INSTANCE_ID, JOB_NAME, JOB_KEY, JOB_PARAMS, CREATE_TIME FROM BATCH_JOB_INSTANCE WHERE JOB_NAME=? AND JOB_KEY=?", jobName, key)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query job instance from db failed", err)
	}
	defer rows.Close()

	for rows.Next() {
		inst := &gobatch.JobInstance{}
		err = rows.Scan(&inst.JobInstanceId, &inst.JobName, &inst.JobKey, &inst.JobParams, &inst.CreateTime)
		if err != nil {
			return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "read job instance failed", err)
		}
		return inst, nil
	}
	return nil, nil
}

func (r *repositoryImpl) FindLastJobInstanceByName(jobName string) (*gobatch.JobInstance, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT JOB_INSTANCE_ID, JOB_NAME, JOB_KEY, JOB_PARAMS, CREATE_TIME FROM BATCH_JOB_INSTANCE WHERE JOB_NAME=? ORDER BY JOB_INSTANCE_ID DESC", jobName)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query job instance from db failed", err)
	}
	defer rows.Close()

	if rows.Next() {
		inst := &gobatch.JobInstance{}
		err = rows.Scan(&inst.JobInstanceId, &inst.JobName, &inst.JobKey, &inst.JobParams, &inst.CreateTime)
		if err != nil {
			return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "read job instance failed", err)
		}
		return inst, nil
	}
	return nil, nil
}

func (r *repositoryImpl) CreateJobInstance(jobName string, jobParams map[string]interface{}) (*gobatch.JobInstance, gobatch.BatchError) {
	str, err := util.JsonString(jobParams)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeGeneral, "jsonify job params failed", err)
	}
	key := util.MD5(str)
	jobInstance := &gobatch.JobInstance{
		JobName:    jobName,
		JobKey:     key,
		JobParams:  str,
		CreateTime: time.Now(),
	}
	res, err := r.db.Exec("INSERT INTO BATCH_JOB_INSTANCE(JOB_NAME, JOB_KEY, JOB_PARAMS, CREATE_TIME) VALUES(?, ?, ?, ?)", jobName, key, str, jobInstance.CreateTime)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "save job instance to db failed", err)
	}
	id, er := res.LastInsertId()
	if er == nil {
		jobInstance.JobInstanceId = id
	}
	return jobInstance, nil
}

// load or save job executions by instance
func (r *repositoryImpl) FindLastJobExecutionByInstance(jobInstance *gobatch.JobInstance) (*gobatch.JobExecution, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT JOB_EXECUTION_ID, JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, START_TIME, END_TIME, STATUS, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION FROM BATCH_JOB_EXECUTION WHERE JOB_INSTANCE_ID=? ORDER BY JOB_EXECUTION_ID DESC", jobInstance.JobInstanceId)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query job executions from db failed", err)
	}
	defer rows.Close()

	jobParams, _ := gobatch.ParseJobParams(jobInstance.JobParams)
	if rows.Next() {
		execution := &BatchJobExecution{}
		err = rows.Scan(&execution.JobExecutionId, &execution.JobInstanceId, &execution.JobName, &execution.CreateTime, &execution.StartTime, &execution.EndTime, &execution.Status, &execution.ExitCode, &execution.ExitMessage, &execution.LastUpdated, &execution.Version)
		if err != nil {
			return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "read job executions failed", err)
		}
		jobExecution := &gobatch.JobExecution{
			JobExecutionId: execution.JobExecutionId,
			JobInstanceId:  execution.JobExecutionId,
			JobName:        execution.JobName,
			JobParams:      jobParams,
			JobStatus:      gobatch.BatchStatus(execution.Status),
			CreateTime:     execution.CreateTime,
			StartTime:      execution.StartTime,
			EndTime:        execution.EndTime,
			Version:        execution.Version,
		}
		return jobExecution, nil
	}
	return nil, nil
}

func (r *repositoryImpl) FindJobExecution(jobExecutionId int64) (*gobatch.JobExecution, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT JOB_EXECUTION_ID, JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, START_TIME, END_TIME, STATUS, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION FROM BATCH_JOB_EXECUTION WHERE JOB_EXECUTION_ID=?", jobExecutionId)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query job executions from db failed", err)
	}
	defer rows.Close()

	if rows.Next() {
		execution := &BatchJobExecution{}
		err = rows.Scan(&execution.JobExecutionId, &execution.JobInstanceId, &execution.JobName, &execution.CreateTime, &execution.StartTime, &execution.EndTime, &execution.Status, &execution.ExitCode, &execution.ExitMessage, &execution.LastUpdated, &execution.Version)
		if err != nil {
			return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "read job executions failed", err)
		}
		jobExecution := &gobatch.JobExecution{
			JobExecutionId: execution.JobExecutionId,
			JobInstanceId:  execution.JobExecutionId,
			JobName:        execution.JobName,
			JobStatus:      gobatch.BatchStatus(execution.Status),
			CreateTime:     execution.CreateTime,
			StartTime:      execution.StartTime,
			EndTime:        execution.EndTime,
			Version:        execution.Version,
		}
		return jobExecution, nil
	}
	return nil, nil
}

func (r *repositoryImpl) SaveJobExecution(execution *gobatch.JobExecution) gobatch.BatchError {
	buff := bytes.NewBufferString("")
	args := make([]interface{}, 0)
	if execution.JobExecutionId == 0 {
		buff.WriteString("INSERT INTO BATCH_JOB_EXECUTION(JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, STATUS, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION) VALUES")
		buff.WriteString("(?, ?, ?, ?, ?, ?, ?, ?)")
		failMsg := ""
		if execution.FailError != nil {
			failMsg = execution.FailError.StackTrace()
		}
		args = append(args, execution.JobInstanceId, execution.JobName, execution.CreateTime, string(execution.JobStatus), execution.JobStatus, failMsg, time.Now(), 1)
		res, err := r.db.Exec(buff.String(), args...)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "save job execution to db failed", err)
		}
		id, _ := res.LastInsertId()
		if id > 0 {
			execution.JobExecutionId = id
		}
		execution.Version = 1
	} else {
		buff.WriteString("UPDATE BATCH_JOB_EXECUTION SET STATUS=?, START_TIME=?, END_TIME=?, EXIT_CODE=?, EXIT_MESSAGE=?, LAST_UPDATED=?, VERSION=? WHERE JOB_EXECUTION_ID=? AND VERSION=?")
		failMsg := ""
		if execution.FailError != nil {
			failMsg = execution.FailError.StackTrace()
		}
		args = append(args, execution.JobStatus, execution.StartTime, execution.EndTime, execution.JobStatus, failMsg, time.Now(), execution.Version+1, execution.JobExecutionId, execution.Version)
		res, err := r.db.Exec(buff.String(), args...)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "update job execution failed", err)
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected <= 0 {
			return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "no job execution in db was updated, job_execution_id:%v version:%v", execution.JobExecutionId, execution.Version)
		}
		execution.Version += 1
	}
	return nil
}

func (r *repositoryImpl) CheckJobStopping(execution *gobatch.JobExecution) (bool, gobatch.BatchError) {
	storeExecution, err := r.FindJobExecution(execution.JobExecutionId)
	if err != nil || storeExecution == nil {
		return false, err
	}
	stoppping := storeExecution.Version != execution.Version && storeExecution.JobStatus == gobatch.STOPPING
	if stoppping {
		execution.Version = storeExecution.Version
	}
	return stoppping, nil
}

func (r *repositoryImpl) FindStepExecutionsByJobExecution(jobExecutionId int64) ([]*gobatch.StepExecution, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT STEP_EXECUTION_ID, STEP_NAME, JOB_EXECUTION_ID, JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, START_TIME, END_TIME, STATUS, COMMIT_COUNT, READ_COUNT, FILTER_COUNT, WRITE_COUNT, READ_SKIP_COUNT, WRITE_SKIP_COUNT, PROCESS_SKIP_COUNT, ROLLBACK_COUNT, EXECUTION_CONTEXT, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION FROM BATCH_STEP_EXECUTION WHERE JOB_EXECUTION_ID=? ORDER BY STEP_EXECUTION_ID DESC", jobExecutionId)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query step executions from db failed", err)
	}
	defer rows.Close()

	results := make([]*gobatch.StepExecution, 0)
	for rows.Next() {
		stepExecution, bError := r.extractStepExecution(rows)
		if bError != nil {
			return nil, bError
		}
		results = append(results, stepExecution)
	}
	return results, nil
}

// find step execution by job_execution and step name
func (r *repositoryImpl) FindStepExecutionsByName(jobExecutionId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT STEP_EXECUTION_ID, STEP_NAME, JOB_EXECUTION_ID, JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, START_TIME, END_TIME, STATUS, COMMIT_COUNT, READ_COUNT, FILTER_COUNT, WRITE_COUNT, READ_SKIP_COUNT, WRITE_SKIP_COUNT, PROCESS_SKIP_COUNT, ROLLBACK_COUNT, EXECUTION_CONTEXT, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION FROM BATCH_STEP_EXECUTION WHERE JOB_EXECUTION_ID=? AND STEP_NAME=? ORDER BY STEP_EXECUTION_ID DESC", jobExecutionId, stepName)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query step executions from db failed", err)
	}
	defer rows.Close()

	if rows.Next() {
		stepExecution, bError := r.extractStepExecution(rows)
		return stepExecution, bError
	}
	return nil, nil
}

// find last step execution by job instance and step name
func (r *repositoryImpl) FindLastStepExecution(jobInstanceId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT STEP_EXECUTION_ID, STEP_NAME, JOB_EXECUTION_ID, JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, START_TIME, END_TIME, STATUS, COMMIT_COUNT, READ_COUNT, FILTER_COUNT, WRITE_COUNT, READ_SKIP_COUNT, WRITE_SKIP_COUNT, PROCESS_SKIP_COUNT, ROLLBACK_COUNT, EXECUTION_CONTEXT, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION FROM BATCH_STEP_EXECUTION WHERE JOB_INSTANCE_ID=? AND STEP_NAME=? ORDER BY CREATE_TIME DESC", jobInstanceId, stepName)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query step executions from db failed", err)
	}
	defer rows.Close()

	if rows.Next() {
		stepExecution, bError := r.extractStepExecution(rows)
		return stepExecution, bError
	}
	return nil, nil
}

// find last step execution by job instance and step name
func (r *repositoryImpl) FindLastCompleteStepExecution(jobInstanceId int64, stepName string) (*gobatch.StepExecution, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT STEP_EXECUTION_ID, STEP_NAME, JOB_EXECUTION_ID, JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, START_TIME, END_TIME, STATUS, COMMIT_COUNT, READ_COUNT, FILTER_COUNT, WRITE_COUNT, READ_SKIP_COUNT, WRITE_SKIP_COUNT, PROCESS_SKIP_COUNT, ROLLBACK_COUNT, EXECUTION_CONTEXT, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION FROM BATCH_STEP_EXECUTION WHERE JOB_INSTANCE_ID=? AND STEP_NAME=? AND STATUS=? ORDER BY CREATE_TIME DESC", jobInstanceId, stepName, gobatch.COMPLETED)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query step executions from db failed", err)
	}
	defer rows.Close()

	if rows.Next() {
		stepExecution, bError := r.extractStepExecution(rows)
		return stepExecution, bError
	}
	return nil, nil
}

func (r *repositoryImpl) extractStepExecution(rows *sql.Rows) (*gobatch.StepExecution, gobatch.BatchError) {
	// 1. query step execution
	execution := &BatchStepExecution{}
	err := rows.Scan(&execution.StepExecutionId, &execution.StepName, &execution.JobExecutionId, &execution.JobInstanceId, &execution.JobName, &execution.CreateTime, &execution.StartTime, &execution.EndTime, &execution.Status, &execution.CommitCount, &execution.ReadCount, &execution.FilterCount, &execution.WriteCount, &execution.ReadSkipCount, &execution.WriteSkipCount, &execution.ProcessSkipCount, &execution.RollbackCount, &execution.ExecutionContext, &execution.ExitCode, &execution.ExitMessage, &execution.LastUpdated, &execution.Version)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "read step executions failed", err)
	}
	// 2. query step context
	batchStepCtx, er := r.FindStepContext(execution.JobInstanceId, execution.StepName)
	if er != nil {
		return nil, er
	}
	if batchStepCtx == nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeGeneral, "can not find step context for step:%v", execution.StepName)
	}
	// 3. construct step execution
	stepContext := gobatch.NewBatchContext()
	if err = util.ParseJson(*batchStepCtx.StepContext, stepContext); err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeGeneral, "parse step context error", err)
	}
	stepExecutionContext := gobatch.NewBatchContext()
	if err = util.ParseJson(execution.ExecutionContext, stepExecutionContext); err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeGeneral, "parse step execution context error", err)
	}
	stepExecution := &gobatch.StepExecution{
		StepExecutionId:      execution.StepExecutionId,
		StepName:             execution.StepName,
		StepStatus:           gobatch.BatchStatus(execution.Status),
		StepContext:          stepContext,
		StepContextId:        batchStepCtx.StepContextId,
		StepExecutionContext: stepExecutionContext,
		CreateTime:           execution.CreateTime,
		StartTime:            execution.StartTime,
		EndTime:              execution.EndTime,
		ReadCount:            execution.ReadCount,
		WriteCount:           execution.WriteCount,
		CommitCount:          execution.CommitCount,
		FilterCount:          execution.FilterCount,
		ReadSkipCount:        execution.ReadSkipCount,
		WriteSkipCount:       execution.WriteSkipCount,
		ProcessSkipCount:     execution.ProcessSkipCount,
		RollbackCount:        execution.RollbackCount,
		LastUpdated:          execution.LastUpdated,
		Version:              execution.Version,
	}
	return stepExecution, nil
}

func (r *repositoryImpl) SaveStepExecution(ctx context.Context, execution *gobatch.StepExecution) gobatch.BatchError {
	buff := bytes.NewBufferString("")
	args := make([]interface{}, 0)
	// 1. save step context
	if execution.StepContextId == 0 {
		stepCtxJson, _ := util.JsonString(execution.StepContext)
		stepContext := &gobatch.StepContext{
			JobInstanceId: execution.JobExecution.JobInstanceId,
			StepName:      execution.StepName,
			StepContext:   &stepCtxJson,
			CreateTime:    execution.CreateTime,
		}
		if err := r.SaveStepContexts(stepContext); err != nil {
			return err
		}
		execution.StepContextId = stepContext.StepContextId
	}
	// 2. save step execution
	if execution.StepExecutionId == 0 {
		buff.WriteString("INSERT INTO BATCH_STEP_EXECUTION(STEP_NAME, JOB_EXECUTION_ID, JOB_INSTANCE_ID, JOB_NAME, CREATE_TIME, STATUS, COMMIT_COUNT, READ_COUNT, FILTER_COUNT, WRITE_COUNT, READ_SKIP_COUNT, WRITE_SKIP_COUNT, PROCESS_SKIP_COUNT, ROLLBACK_COUNT, EXECUTION_CONTEXT, STEP_CONTEXT_ID, EXIT_CODE, EXIT_MESSAGE, LAST_UPDATED, VERSION) VALUES")
		buff.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		executionCtxJson, err := util.JsonString(execution.StepExecutionContext)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "jsonify step execution context failed", err)
		}
		failMsg := ""
		if execution.FailError != nil {
			failMsg = execution.FailError.StackTrace()
		}
		args = append(args, execution.StepName, execution.JobExecution.JobExecutionId, execution.JobExecution.JobInstanceId, execution.JobExecution.JobName, execution.CreateTime, execution.StepStatus, execution.CommitCount, execution.ReadCount, execution.FilterCount, execution.WriteCount, execution.ReadSkipCount, execution.WriteSkipCount, execution.ProcessSkipCount, execution.RollbackCount, executionCtxJson, execution.StepContextId, execution.StepStatus, failMsg, time.Now(), 1)

		res, err := r.db.Exec(buff.String(), args...)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "save step execution to db failed", err)
		}
		// 3. fill back step execution
		id, _ := res.LastInsertId()
		if id > 0 {
			execution.StepExecutionId = id
		}
		execution.Version = 1
	} else {
		executionCtxStr, _ := util.JsonString(execution.StepExecutionContext)
		buff.WriteString("UPDATE BATCH_STEP_EXECUTION SET STATUS=?, COMMIT_COUNT=?, READ_COUNT=?, FILTER_COUNT=?, WRITE_COUNT=?, READ_SKIP_COUNT=?, WRITE_SKIP_COUNT=?, PROCESS_SKIP_COUNT=?, ROLLBACK_COUNT=?, EXECUTION_CONTEXT=?, STEP_CONTEXT_ID=?, EXIT_CODE=?, EXIT_MESSAGE=?, START_TIME=?, END_TIME=?, LAST_UPDATED=?, VERSION=? WHERE STEP_EXECUTION_ID=? AND VERSION=?")
		failMsg := ""
		if execution.FailError != nil {
			failMsg = execution.FailError.StackTrace()
		}
		args = append(args, execution.StepStatus, execution.CommitCount, execution.ReadCount, execution.FilterCount, execution.WriteCount, execution.ReadSkipCount, execution.WriteSkipCount, execution.ProcessSkipCount, execution.RollbackCount, executionCtxStr, execution.StepContextId, execution.StepStatus, failMsg, execution.StartTime, execution.EndTime, time.Now(), execution.Version+1, execution.StepExecutionId, execution.Version)

		res, err := r.db.Exec(buff.String(), args...)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "update step execution failed", err)
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected <= 0 {
			return gobatch.NewBatchError(gobatch.ErrCodeConcurrency, "no step execution in db was updated, step_execution_id:%v version:%v", execution.StepExecutionId, execution.Version)
		}
		execution.Version += 1
	}
	// 3. check job status
	stopping, err := r.CheckJobStopping(execution.JobExecution)
	if err != nil {
		return err
	}
	if stopping && execution.StepStatus != gobatch.STOPPED {
		if execution.StepStatus == gobatch.STARTING || execution.StepStatus == gobatch.STARTED {
			oldStatus := execution.StepStatus
			execution.StepStatus = gobatch.STOPPED
			execution.EndTime = time.Now()
			err = r.UpdateStepStatus(execution)
			i := 0
			for err != nil && i < 3 {
				err = r.UpdateStepStatus(execution)
				r.logger.Warn(ctx, "update step status to STOPPED failed for %d times, step:%v, err:%v", i, execution.StepName, err)
			}
			if err != nil {
				execution.StepStatus = oldStatus
				r.logger.Error(ctx, "update step status to STOPPED failed, step:%v, err:%v", execution.StepName, err)
			}
		}
		return gobatch.NewBatchError(gobatch.ErrCodeStop, "the job is stopped")
	}
	return nil
}

func (r *repositoryImpl) UpdateStepStatus(execution *gobatch.StepExecution) gobatch.BatchError {
	buff := bytes.NewBufferString("")
	args := make([]interface{}, 0)
	buff.WriteString("UPDATE BATCH_STEP_EXECUTION SET STATUS=?, VERSION=? WHERE STEP_EXECUTION_ID=? AND VERSION=?")
	args = append(args, execution.StepStatus, execution.Version+1, execution.StepExecutionId, execution.Version)

	res, err := r.db.Exec(buff.String(), args...)
	if err != nil {
		return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "update step status failed", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected <= 0 {
		return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "no step execution in db was updated, step_execution_id:%v, version:%v", execution.StepContextId, execution.Version)
	}
	execution.Version += 1
	return nil
}

// load or save step executions by instance

// load or save step context
func (r *repositoryImpl) FindStepContext(jobInstanceId int64, stepName string) (*gobatch.StepContext, gobatch.BatchError) {
	rows, err := r.db.Query("SELECT STEP_CONTEXT_ID, JOB_INSTANCE_ID, STEP_NAME, STEP_CONTEXT, CREATE_TIME FROM BATCH_STEP_CONTEXT WHERE JOB_INSTANCE_ID=? AND STEP_NAME=?", jobInstanceId, stepName)
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "query step context from db failed", err)
	}
	defer rows.Close()

	if rows.Next() {
		stepCtx := &gobatch.StepContext{}
		err = rows.Scan(&stepCtx.StepContextId, &stepCtx.JobInstanceId, &stepCtx.StepName, &stepCtx.StepContext, &stepCtx.CreateTime)
		if err != nil {
			return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "read step context failed", err)
		}
		return stepCtx, nil
	}
	return nil, nil
}

func (r *repositoryImpl) SaveStepContexts(stepCtx *gobatch.StepContext) gobatch.BatchError {
	buff := bytes.NewBufferString("")
	args := make([]interface{}, 0)
	if stepCtx.StepContextId == 0 {
		buff.WriteString("INSERT INTO BATCH_STEP_CONTEXT(JOB_INSTANCE_ID, STEP_NAME, STEP_CONTEXT, CREATE_TIME) VALUES")
		buff.WriteString("(?, ?, ?, ?)")
		args = append(args, stepCtx.JobInstanceId, stepCtx.StepName, stepCtx.StepContext, stepCtx.CreateTime)
		res, err := r.db.Exec(buff.String(), args...)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "save step context to db failed", err)
		}
		id, er := res.LastInsertId()
		if er == nil {
			stepCtx.StepContextId = id
		}
		buff.Reset()
		args = args[0:0]
	}
	return nil
}
