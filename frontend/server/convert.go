package server

import (
	"encoding/json"
	"errors"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/frontend/proto"
	"github.com/lijiang2014/cwl.go/intergration/slex"
	"github.com/lijiang2014/cwl.go/message"
	"starlight/common/model"
)

// FromGrpcStatus generate an array from a grpc struct proto.Status
func FromGrpcStatus(raw *proto.Status) (*message.StepStatusArray, error) {
	ret := &message.StepStatusArray{}

	// deep copy
	for _, ss := range raw.Steps {
		tmp := message.StepStatus{
			ID:     message.NewPathID(ss.Path),
			Status: message.MessageStatus(ss.Status),
			JobID:  ss.Job,
		}

		tmp.Output = cwl.Values{}
		if ss.Values != nil {
			err := json.Unmarshal([]byte(*ss.Values), &tmp.Output)
			if err != nil {
				return nil, err
			}
		}

		if ss.Error != nil {
			tmp.Error = errors.New(*ss.Error)
		}
		ret.Append(&tmp)
	}
	return ret, nil
}

func ToGrpcStepStatus(ss *message.StepStatus) *proto.StepStatus {
	tmp, _ := json.Marshal(ss.Output)
	tmpStr := string(tmp)
	ret := proto.StepStatus{
		Path:   ss.ID.Path(),
		Status: string(ss.Status),
		Job:    ss.JobID,
		Values: &tmpStr,
	}
	if ss.Error != nil {
		tmp := ss.Error.Error()
		ret.Error = &tmp
	}
	return &ret
}

func ToGrpcStatus(status *message.StepStatusArray) *proto.Status {
	var tmp []*proto.StepStatus
	status.Foreach(func(s *message.StepStatus) {
		tmp = append(tmp, ToGrpcStepStatus(s))
	})
	return &proto.Status{
		Result: &proto.Result{Success: true},
		Steps:  tmp,
	}
}

func FromGrpcAllocation(g *proto.Allocation) *slex.JobAllocationModel {
	ret := slex.JobAllocationModel{
		Default: FromGrpcSingleAllocation(g.Default),
		Diff:    map[string]*slex.SingleJobAllocationModel{},
	}
	for k, v := range g.Diff {
		ret.Diff[k] = FromGrpcSingleAllocation(v)
	}
	return &ret
}

func FromGrpcSingleAllocation(g *proto.SingleAllocation) *slex.SingleJobAllocationModel {
	ret := slex.SingleJobAllocationModel{
		Cluster:   g.Cluster,
		Partition: g.Partition,
		Cpu:       CopyInt32Pointer(g.Cpu),
		Gpu:       CopyInt32Pointer(g.Gpu),
		Memory:    CopyInt64Pointer(g.Memory),
		WorkDir:   model.Volume{},
	}
	if g.Workdir != nil {
		ret.WorkDir = model.Volume{HostPath: *g.Workdir}
	}
	return &ret
}

func CopyInt32Pointer(i32 *int32) *int {
	if i32 == nil {
		return nil
	}
	var tmp int = int(*i32)
	return &tmp
}

func CopyInt64Pointer(i64 *int64) *int {
	if i64 == nil {
		return nil
	}
	var tmp int = int(*i64)
	return &tmp
}
