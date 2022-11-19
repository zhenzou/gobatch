package txn

import (
	"database/sql"

	"github.com/chararch/gobatch"
)

// DefaultTxManager default TransactionManager implementation
type DefaultTxManager struct {
	db *sql.DB
}

// NewTransactionManager create a TransactionManager instance
func NewTransactionManager(db *sql.DB) gobatch.TransactionManager {
	return &DefaultTxManager{
		db: db,
	}
}

// BeginTx begin a transaction
func (tm *DefaultTxManager) BeginTx() (interface{}, gobatch.BatchError) {
	tx, err := tm.db.Begin()
	if err != nil {
		return nil, gobatch.NewBatchError(gobatch.ErrCodeDbFail, "start transaction failed", err)
	}
	return tx, nil
}

// Commit commit a transaction
func (tm *DefaultTxManager) Commit(tx interface{}) gobatch.BatchError {
	tx1 := tx.(*sql.Tx)
	err := tx1.Commit()
	if err != nil {
		return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "transaction commit failed", err)
	}
	return nil
}

// Rollback rollback a transaction
func (tm *DefaultTxManager) Rollback(tx interface{}) gobatch.BatchError {
	tx1 := tx.(*sql.Tx)
	err := tx1.Rollback()
	if err != nil {
		return gobatch.NewBatchError(gobatch.ErrCodeDbFail, "transaction rollback failed", err)
	}
	return nil
}
