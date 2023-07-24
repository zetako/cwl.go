package slex

import (
	"context"
	"github.com/google/uuid"
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

func New(ctx context.Context, token string, username string, alloc *JobAllocationModel) (*StarlightExecutor, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	c, err := httpclient.NewBihuClient(ctx, token)
	if err != nil {
		return nil, err
	}
	ret := StarlightExecutor{
		alloc:    alloc,
		ctx:      ctx,
		token:    token,
		username: username,
		uuid:     id,
		client:   c,
	}

	return &ret, nil
}
