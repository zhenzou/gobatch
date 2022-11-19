package gobatch

import (
	"os"
)

var DefaultLogger Logger

// SetLogger set a logger instance for GoBatch
func SetLogger(logger Logger) {
	DefaultLogger = logger
}

func init() {
	DefaultLogger = NewLogger(os.Stdout, Info)

}

// task pool
const (
	DefaultJobPoolSize      = 10
	DefaultStepTaskPoolSize = 1000
)

var jobPool = newTaskPool(DefaultJobPoolSize)
var stepPool = newTaskPool(DefaultStepTaskPoolSize)

// SetMaxRunningJobs set max number of parallel jobs for GoBatch
func SetMaxRunningJobs(size int) {
	jobPool.SetMaxSize(size)
}

// SetMaxRunningSteps set max number of parallel steps for GoBatch
func SetMaxRunningSteps(size int) {
	stepPool.SetMaxSize(size)
}
