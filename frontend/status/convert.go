package status

import (
	"encoding/json"
	"errors"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/frontend/proto"
	"github.com/lijiang2014/cwl.go/runner/message"
)

// FromGrpcStatus generate an array from a grpc struct proto.Status
//
// TODO Should NOT BE an StepStatusArray method
func (a *StepStatusArray) FromGrpcStatus(raw *proto.Status) error {
	// lock
	a.Lock()
	defer a.Unlock()

	// alloc
	a.array = []*StepStatus{}

	// deep copy
	for _, ss := range raw.Steps {
		tmp := StepStatus{
			ID:     message.NewPathID(ss.Path),
			Status: message.MessageStatus(ss.Status),
			JobID:  ss.Job,
		}

		tmp.Output = cwl.Values{}
		if ss.Values != nil {
			err := json.Unmarshal([]byte(*ss.Values), &tmp.Output)
			if err != nil {
				return err
			}
		}

		if ss.Error != nil {
			tmp.Error = errors.New(*ss.Error)
		}
		// don't use locked Append here, will cause dead lock
		//a.Append(&tmp)
		a.array = append(a.array, &tmp)
	}
	return nil
}

func ToGrpcStepStatus(ss *StepStatus) *proto.StepStatus {
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

func ToGrpcStatus(status *StepStatusArray) *proto.Status {
	var tmp []*proto.StepStatus
	status.Foreach(func(s *StepStatus) {
		tmp = append(tmp, ToGrpcStepStatus(s))
	})
	return &proto.Status{
		Result: &proto.Result{Success: true},
		Steps:  tmp,
	}
}
