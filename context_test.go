package gobatch

import (
	"fmt"
	"testing"

	"github.com/bmizerany/assert"
)

func TestBatchContext_Get(t *testing.T) {
	ctx := NewBatchContext()
	v := ctx.Get("key")
	assert.Equal(t, v, nil)

	ctx.Put("key", "1111")
	assert.Equal(t, ctx.Get("key"), "1111")
}

type Key struct {
	Id   int64
	Code string
}

func TestBatchContext_MarshalJSON(t *testing.T) {
	batchCtx := NewBatchContext()
	batchCtx.Put("count", 100)
	batchCtx.Put("current", 5)
	batchCtx.Put("keys", []Key{{
		Id:   1,
		Code: "1",
	}, {
		Id:   2,
		Code: "2",
	}, {
		Id:   3,
		Code: "3",
	},
	})
	json := batchCtx.ToString()
	fmt.Printf("json:%v\n", json)

	batchCtx2 := NewBatchContext()
	err := batchCtx2.FromString(json)
	assert.Equal(t, nil, err)
	fmt.Printf("batchCtx:%+v\n", batchCtx)
	fmt.Printf("batchCtx2:%+v\n", batchCtx2)
}
