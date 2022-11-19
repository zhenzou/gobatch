package files

import (
	"context"
	"io"

	"github.com/chararch/gobatch"
)

func NewCopier(filesToMove ...FileMove) gobatch.Handler {
	return &_FileCopyHandler{filesToMove: filesToMove}
}

type _FileCopyHandler struct {
	filesToMove []FileMove
}

func (handler *_FileCopyHandler) Handle(execution *gobatch.StepExecution) gobatch.BatchError {
	for _, fm := range handler.filesToMove {
		ffp := &FilePath{fm.FromFileName}
		fromFileName, err := ffp.Format(execution)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "get real file path:%v err", fm.FromFileName, err)
		}
		tfp := &FilePath{fm.ToFileName}
		toFileName, err := tfp.Format(execution)
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "get real file path:%v err", fm.ToFileName, err)
		}

		// open from-file
		ffs := fm.FromFileStore
		reader, err := ffs.Open(fromFileName, "utf-8")
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "open from file:%v err", fromFileName, err)
		}

		// create to-file
		tfs := fm.ToFileStore
		writer, err := tfs.Create(toFileName, "utf-8")
		if err != nil {
			if er := reader.Close(); er != nil {
				gobatch.DefaultLogger.Error(context.Background(), "close file reader:%v error", fromFileName, er)
			}
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "open to file:%v err", toFileName, err)
		}

		_, err = io.Copy(writer, reader)

		if er := reader.Close(); er != nil {
			gobatch.DefaultLogger.Error(context.Background(), "close file writer:%v error", toFileName, er)
		}
		if er := writer.Close(); er != nil {
			gobatch.DefaultLogger.Error(context.Background(), "close file writer:%v error", toFileName, er)
		}
		if err != nil {
			return gobatch.NewBatchError(gobatch.ErrCodeGeneral, "copy file: %v -> %v error", fromFileName, toFileName, err)
		}
	}
	return nil

}
