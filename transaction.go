package gobatch

// TransactionManager used by chunk step to execute chunk process in a transaction.
type TransactionManager interface {
	BeginTx() (tx interface{}, err BatchError)
	Commit(tx interface{}) BatchError
	Rollback(tx interface{}) BatchError
}
