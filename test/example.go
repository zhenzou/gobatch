package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/chararch/gobatch"
	"github.com/chararch/gobatch/adapters/repository"
	"github.com/chararch/gobatch/util"
)

// simple task
func mytask() {
	fmt.Println("mytask executed")
}

// reader
type myReader struct {
}

func (r *myReader) Read(chunkCtx *gobatch.ChunkContext) (interface{}, gobatch.BatchError) {
	curr, _ := chunkCtx.StepExecution.StepContext.GetInt("read.num", 0)
	if curr < 100 {
		chunkCtx.StepExecution.StepContext.Put("read.num", curr+1)
		return fmt.Sprintf("value-%v", curr), nil
	}
	return nil, nil
}

// processor
type myProcessor struct{}

func (r *myProcessor) Process(item interface{}, chunkCtx *gobatch.ChunkContext) (interface{}, gobatch.BatchError) {
	return fmt.Sprintf("processed-%v", item), nil
}

// writer
type myWriter struct{}

func (r *myWriter) Write(items []interface{}, chunkCtx *gobatch.ChunkContext) gobatch.BatchError {
	fmt.Printf("write: %v\n", items)
	return nil
}

func main() {
	// set db for gobatch to store job&step execution context
	var db *sql.DB
	var err error
	db, err = sql.Open("mysql", "root:root123@tcp(127.0.0.1:3306)/gobatch?charset=utf8&parseTime=true")
	if err != nil {
		panic(err)
	}
	logger := gobatch.NewLogger(os.Stdout, gobatch.Info)

	repo := repository.New(db, logger)
	engine := gobatch.NewEngine(repo)

	txnMgr := gobatch.NewTransactionManager(db)

	stepFactory := gobatch.NewStepBuilderFactory(repo, txnMgr)
	// build steps
	step1 := stepFactory.Get("mytask").Handler(mytask).Build()
	// step2 := gobatch.NewStep("my_step").Handler(&myReader{}, &myProcessor{}, &myWriter{}).Build()
	step2 := stepFactory.Get("my_step").Reader(&myReader{}).Processor(&myProcessor{}).Writer(&myWriter{}).ChunkSize(10).Build()

	// build job
	jobFactory := gobatch.NewJobBuilderFactory(repo)
	job := jobFactory.Get("my_job").Start(step1).Next(step2).Build()

	// register job to gobatch
	err = engine.Register(job)
	if err != nil {
		panic(err)
	}

	// run
	// gobatch.StartAsync(context.Background(), job.Name(), "")
	params, _ := util.JsonString(map[string]interface{}{
		"rand": time.Now().Nanosecond(),
	})
	_, err = engine.Start(context.Background(), job.Name(), params)
	if err != nil {
		panic(err)
	}
}
