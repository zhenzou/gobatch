package files

import (
	"fmt"
	"strings"

	"github.com/chararch/gobatch"
)

const (
	fileItemWriterHandleKey   = "gobatch.FileItemWriter.handle"
	fileItemWriterFileNameKey = "gobatch.FileItemWriter.fileName"
)

func NewWriter(fd FileObjectModel, writers ...interface{}) gobatch.Writer {
	fw := &_FileWriter{fd: fd}
	if len(writers) > 0 {
		for _, w := range writers {
			switch ww := w.(type) {
			case FileItemWriter:
				fw.writer = ww
			case ChecksumFlusher:
				fw.checkumer = ww
			case FileMerger:
				fw.merger = ww
			}
		}
	}
	if fw.writer == nil && fw.fd.Type != "" {
		fw.writer = GetFileItemWriter(fw.fd.Type)
	}
	if fw.writer == nil {
		panic("file type is non-standard and no FileItemWriter specified")
	}
	if fw.merger == nil && fw.fd.Type != "" {
		fw.merger = GetFileMergeSplitter(fw.fd.Type)
	}
	return fw
}

type _FileWriter struct {
	fd        FileObjectModel
	writer    FileItemWriter
	checkumer ChecksumFlusher
	merger    FileMerger
}

func (w *_FileWriter) Open(execution *gobatch.StepExecution) gobatch.BatchError {
	stepName := execution.StepName
	// get actual file name
	fd := w.fd // copy fd
	fp := &FilePath{fd.FileName}
	fileName, err := fp.Format(execution)
	if err != nil {
		return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "get real file path:%v err", fd.FileName, err)
	}
	fd.FileName = fileName
	if strings.Index(stepName, ":") > 0 { // may be a partitioned step
		fd.FileName = fmt.Sprintf("%v.%v", fd.FileName, strings.ReplaceAll(stepName, ":", "."))
	}
	handle, e := w.writer.Open(fd)
	if e != nil {
		return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "open file writer:%v err", fd.FileName, e)
	}
	execution.StepExecutionContext.Put(fileItemWriterHandleKey, handle)
	execution.StepExecutionContext.Put(fileItemWriterFileNameKey, fd.FileName)
	return nil
}
func (w *_FileWriter) Write(items []interface{}, chunkCtx *gobatch.ChunkContext) gobatch.BatchError {
	executionCtx := chunkCtx.StepExecution.StepExecutionContext
	handle := executionCtx.Get(fileItemWriterHandleKey)
	fileName := executionCtx.Get(fileItemWriterFileNameKey)
	for _, item := range items {
		e := w.writer.WriteItem(handle, item)
		if e != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "write item to file:%v err", fileName, e)
		}
	}
	return nil
}
func (w *_FileWriter) Close(execution *gobatch.StepExecution) gobatch.BatchError {
	executionCtx := execution.StepExecutionContext
	handle := executionCtx.Get(fileItemWriterHandleKey)
	fileName := executionCtx.Get(fileItemWriterFileNameKey)
	executionCtx.Remove(fileItemWriterHandleKey)
	e := w.writer.Close(handle)
	if e != nil {
		return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "close file writer:%v err", fileName, e)
	}
	// generate file checksum
	if w.fd.Checksum != "" {
		fd := w.fd
		fd.FileName = fileName.(string)
		checksumer := GetChecksumer(fd.Checksum)
		if checksumer != nil {
			err := checksumer.Checksum(fd)
			if err != nil {
				return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "generate file checksum:%v err", fd, err)
			}
		}
	}
	return nil
}

func (w *_FileWriter) Aggregate(execution *gobatch.StepExecution, subExecutions []*gobatch.StepExecution) gobatch.BatchError {
	if w.merger != nil {
		subFiles := make([]FileObjectModel, 0)
		for _, subExecution := range subExecutions {
			fileName := subExecution.StepExecutionContext.Get(fileItemWriterFileNameKey)
			fd := w.fd
			fd.FileName = fileName.(string)
			subFiles = append(subFiles, fd)
		}
		fd := w.fd
		fp := &FilePath{fd.FileName}
		fileName, err := fp.Format(execution)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "get real file path:%v err", fd.FileName, err)
		}
		fd.FileName = fileName
		err = w.merger.Merge(subFiles, fd)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "aggregate file:%v err", fd.FileName, err)
		}
		// generate file checksum
		if fd.Checksum != "" {
			checksumer := GetChecksumer(fd.Checksum)
			if checksumer != nil {
				err = checksumer.Checksum(fd)
				if err != nil {
					return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "generate file checksum:%v err", fd, err)
				}
			}
		}
	}
	return nil
}
