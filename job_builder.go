package gobatch

import (
	"fmt"
)

type JobBuilderFactory interface {
	Get(name string) JobBuilder
}

func NewJobBuilderFactory(repository Repository) JobBuilderFactory {
	return &jobBuilderFactory{
		repository: repository,
		builders:   map[string]JobBuilder{},
	}
}

type jobBuilderFactory struct {
	repository Repository
	builders   map[string]JobBuilder
}

func (j jobBuilderFactory) Get(name string) JobBuilder {
	return &jobBuilder{
		name: name,
	}
}

type JobBuilder interface {
	Start(step Step) SimpleJobBuilder
}

type jobBuilder struct {
	name       string
	repository Repository
}

func (builder *jobBuilder) Start(step Step) SimpleJobBuilder {
	return newSimpleJobBuilder(builder.name, builder.repository, step)
}

type SimpleJobBuilder interface {
	Next(step Step) SimpleJobBuilder
	Repository(repository Repository) SimpleJobBuilder
	Steps(step ...Step) SimpleJobBuilder
	Listener(listener ...interface{}) SimpleJobBuilder
	// Build TODO params
	Build() Job
}

type simpleJobBuilder struct {
	name               string
	steps              []Step
	jobListeners       []JobListener
	stepListeners      []StepListener
	chunkListeners     []ChunkListener
	partitionListeners []PartitionListener
	repository         Repository
}

func newSimpleJobBuilder(name string, repository Repository, step Step) SimpleJobBuilder {
	return &simpleJobBuilder{
		name:       name,
		steps:      []Step{step},
		repository: repository,
	}
}

func (builder *simpleJobBuilder) Next(step Step) SimpleJobBuilder {
	builder.steps = append(builder.steps, step)
	return builder
}

func (builder *simpleJobBuilder) Repository(repository Repository) SimpleJobBuilder {
	builder.repository = repository
	return builder
}

func (builder *simpleJobBuilder) Steps(step ...Step) SimpleJobBuilder {
	builder.steps = append(builder.steps, step...)
	return builder
}

func (builder *simpleJobBuilder) Listener(listener ...interface{}) SimpleJobBuilder {
	for _, l := range listener {
		switch ll := l.(type) {
		case JobListener:
			builder.jobListeners = append(builder.jobListeners, ll)
		case StepListener:
			builder.stepListeners = append(builder.stepListeners, ll)
		case ChunkListener:
			builder.chunkListeners = append(builder.chunkListeners, ll)
		case PartitionListener:
			builder.partitionListeners = append(builder.partitionListeners, ll)
		default:
			panic(fmt.Sprintf("not supported listener:%+v for job:%v", ll, builder.name))
		}
	}
	return builder
}

func (builder *simpleJobBuilder) Build() Job {
	var job Job
	for _, sl := range builder.stepListeners {
		for _, step := range builder.steps {
			step.addListener(sl)
		}
	}
	for _, cl := range builder.chunkListeners {
		for _, step := range builder.steps {
			if chkStep, ok := step.(*chunkStep); ok {
				chkStep.addChunkListener(cl)
			}
		}
	}
	for _, pl := range builder.partitionListeners {
		for _, step := range builder.steps {
			if chkStep, ok := step.(*partitionStep); ok {
				chkStep.addPartitionListener(pl)
			}
		}
	}
	job = newSimpleJob(builder.name, builder.steps, builder.jobListeners, builder.repository)
	return job
}
