package example2

import (
	"database/sql"
	"time"

	"github.com/chararch/gobatch"
	file2 "github.com/chararch/gobatch/extensions/file"
)

var tradeFile = file2.FileObjectModel{
	FileStore:     &file2.LocalFileSystem{},
	FileName:      "res/trade.data",
	Type:          file2.TSV,
	Encoding:      "utf-8",
	Header:        false,
	ItemPrototype: &Trade{},
}

var tradeFileExport = file2.FileObjectModel{
	FileStore:     &file2.LocalFileSystem{},
	FileName:      "res/{date,yyyyMMdd}/trade.csv",
	Type:          file2.CSV,
	Encoding:      "utf-8",
	Checksum:      file2.MD5,
	ItemPrototype: &Trade{},
}

var ftp = &file2.FTPFileSystem{
	Hort:        "localhost",
	Port:        21,
	User:        "gobatch",
	Password:    "gobatch123",
	ConnTimeout: time.Second,
}

var copyFileToFtp = file2.FileMove{
	FromFileName:  "res/{date,yyyyMMdd}/trade.csv",
	FromFileStore: &file2.LocalFileSystem{},
	ToFileStore:   ftp,
	ToFileName:    "trade/{date,yyyyMMdd}/trade.csv",
}
var copyChecksumFileToFtp = file2.FileMove{
	FromFileName:  "res/{date,yyyyMMdd}/trade.csv.md5",
	FromFileStore: &file2.LocalFileSystem{},
	ToFileStore:   ftp,
	ToFileName:    "trade/{date,yyyyMMdd}/trade.csv.md5",
}

type tradeImporter struct {
	db *sql.DB
}

func (p *tradeImporter) Write(items []interface{}, chunkCtx *gobatch.ChunkContext) gobatch.BatchError {
	for _, item := range items {
		trade := item.(*Trade)
		_, err := p.db.Exec("INSERT INTO t_trade(trade_no, account_no, type, amount, terms, interest_rate, trade_time, status) values (?,?,?,?,?,?,?,?)",
			trade.TradeNo, trade.AccountNo, trade.Type, trade.Amount, trade.Terms, trade.InterestRate, trade.TradeTime, trade.Status)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "insert trade into db err", err)
		}
	}
	return nil
}
