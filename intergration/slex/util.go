package slex

import (
	"context"
	"github.com/lijiang2014/cwl.go/intergration/client"
	"path"
)

// CopyIntPointer 复制一个int指针的内容，并作为另一个int指针返回
func CopyIntPointer(src *int) (dst *int) {
	if src == nil {
		return nil
	}

	var tmp int
	tmp = *src
	return &tmp
}

// New 创建一个新的slex实例
func New(ctx context.Context, id string, c *client.StarlightClient, username string, alloc *JobAllocationModel) (*StarlightExecutor, error) {
	AddWorkdirSuffix(alloc, id)
	ret := StarlightExecutor{
		alloc:      alloc,
		ctx:        ctx,
		username:   username,
		workflowID: id,
		client:     c,
	}

	return &ret, nil
}

// AddWorkdirSuffix 修改工作目录为一个指定的子目录，子目录名是执行器的id
func AddWorkdirSuffix(alloc *JobAllocationModel, id string) {
	if alloc.Default.WorkDir.HostPath != "" {
		alloc.Default.WorkDir.HostPath = path.Join(alloc.Default.WorkDir.HostPath, id)
	}
	for k, v := range alloc.Diff {
		if v.WorkDir.HostPath != "" {
			alloc.Diff[k].WorkDir.HostPath = path.Join(v.WorkDir.HostPath, id)
		}
	}
}
