package gobatch

type Task func(execution *StepExecution) BatchError

type Handler interface {
	Handle(execution *StepExecution) BatchError
}

type Reader interface {
	//Read each call of Read() will return a data item, if there is no more data, a nil item will be returned.
	Read(chunkCtx *ChunkContext) (interface{}, BatchError)
}
type Processor interface {
	//Process process an item from reader and return a result item
	Process(item interface{}, chunkCtx *ChunkContext) (interface{}, BatchError)
}
type Writer interface {
	//Write write items generated by processor in a chunk
	Write(items []interface{}, chunkCtx *ChunkContext) BatchError
}

type Partitioner interface {
	//Partition generate sub step executions from specified step execution and partitions count
	Partition(execution *StepExecution, partitions uint) ([]*StepExecution, BatchError)
	//GetPartitionNames generate sub step names from specified step execution and partitions count
	GetPartitionNames(execution *StepExecution, partitions uint) []string
}
type PartitionerFactory interface {
	GetPartitioner(minPartitionSize, maxPartitionSize uint) Partitioner
}

type Aggregator interface {
	//Aggregate aggregate result from all sub step executions
	Aggregate(execution *StepExecution, subExecutions []*StepExecution) BatchError
}