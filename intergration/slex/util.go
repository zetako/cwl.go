package slex

import (
	"context"
	"path"
	"starlight/common/httpclient"
)

func CopyIntPointer(src *int) (dst *int) {
	if src == nil {
		return nil
	}

	var tmp int
	tmp = *src
	return &tmp
}

func New(ctx context.Context, id string, token string, username string, alloc *JobAllocationModel) (*StarlightExecutor, error) {
	c, err := httpclient.NewBihuClient(ctx, token)
	if err != nil {
		return nil, err
	}
	AddWorkdirSuffix(alloc, id)
	ret := StarlightExecutor{
		alloc:      alloc,
		ctx:        ctx,
		token:      token,
		username:   username,
		workflowID: id,
		client:     c,
	}

	return &ret, nil
}

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
