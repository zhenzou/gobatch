package example2

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/chararch/gobatch"
	"github.com/chararch/gobatch/adapters/logger"
	"github.com/chararch/gobatch/adapters/repository"
	"github.com/chararch/gobatch/util"
)

func openDB() *sql.DB {
	var sqlDb *sql.DB
	var err error
	sqlDb, err = sql.Open("mysql", "root:root123@tcp(127.0.0.1:3306)/example?charset=utf8&parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	return sqlDb
}

func removeJobData() {
	sqlDb := openDB()
	_, err := sqlDb.Exec("DELETE FROM t_trade")
	if err != nil {
		log.Fatal(err)
	}
	_, err = sqlDb.Exec("DELETE FROM t_repay_plan")
	if err != nil {
		log.Fatal(err)
	}
}

func buildAndRunJob() {
	sqlDb := openDB()

	logger := logger.NewLogger(os.Stdout, logger.Info)
	repo := repository.New(sqlDb, logger)
	engine := gobatch.NewEngine(repo)

	txnMgr := gobatch.NewTransactionManager(sqlDb)
	gobatch.SetTransactionManager(txnMgr)

	stepFactory := gobatch.NewStepBuilderFactory(repo, txnMgr)

	step1 := stepFactory.Get("import_trade").ReadFile(tradeFile).Writer(&tradeImporter{sqlDb}).Partitions(10).Build()
	step2 := stepFactory.Get("gen_repay_plan").Reader(&tradeReader{sqlDb}).Handler(&repayPlanHandler{sqlDb}).Partitions(10).Build()
	step3 := stepFactory.Get("stats").Handler(&statsHandler{sqlDb}).Build()
	step4 := stepFactory.Get("export_trade").Reader(&tradeReader{sqlDb}).WriteFile(tradeFileExport).Partitions(10).Build()
	step5 := stepFactory.Get("upload_file_to_ftp").CopyFile(copyFileToFtp, copyChecksumFileToFtp).Build()

	factory := gobatch.NewJobBuilderFactory(repo)

	job := factory.Get("accounting_job").Start(step1).Steps(step2, step3, step4, step5).Build()

	engine.Register(job)

	params, _ := util.JsonString(map[string]interface{}{
		"date": time.Now().Format("2006-01-02"),
		"rand": time.Now().Nanosecond(),
	})
	engine.Start(context.Background(), job.Name(), params)
}
