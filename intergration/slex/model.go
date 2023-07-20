package slex

import (
	"github.com/lijiang2014/cwl.go/message"
	"starlight/common/model"
)

type JobSubmitModel struct {
	RuntimeParams struct {
		// Basic Params
		JobName   string            `json:"jobName"`
		Cluster   string            `json:"cluster"`
		Partition string            `json:"partition"`
		Env       map[string]string `json:"env,omitempty"`
		Endpoints []model.Proxy     `json:"endpoints,omitempty"`
		Stdin     string            `json:"stdin,omitempty"`
		Stdout    string            `json:"stdout,omitempty"`
		Stderr    string            `json:"stderr,omitempty"`
		// Direct Task specified
		WorkDir model.Volume `json:"workDir,omitempty"`
		Cmd     []string     `json:"cmd"`
		// Container Task specified
		Image  string         `json:"image,omitempty"`
		Volume []model.Volume `json:"volume,omitempty"`
		//Kind   string         `json:"kind,omitempty"` // Kind not needed, for workflow, always job
		// Runtime Requirements
		Cpu    int `json:"cpu,omitempty"`
		Gpu    int `json:"gpu,omitempty"`
		Memory int `json:"memory,omitempty"`
	} `json:"runtime_params"`
}

type SingleJobAllocationModel struct {
	Cluster   string       `json:"cluster"`
	Partition string       `json:"partition"`
	Cpu       *int         `json:"cpu,omitempty"`
	Gpu       *int         `json:"gpu,omitempty"`
	Memory    *int         `json:"memory,omitempty"`
	WorkDir   model.Volume `json:"workDir,omitempty"`
}

func (s *SingleJobAllocationModel) Copy() SingleJobAllocationModel {
	ret := SingleJobAllocationModel{
		Cluster:   s.Cluster,
		Partition: s.Partition,
	}
	if s.Cpu != nil {
		tmpC := *(s.Cpu)
		ret.Cpu = &tmpC
	}
	if s.Gpu != nil {
		tmpG := *(s.Gpu)
		ret.Gpu = &tmpG
	}
	if s.Memory != nil {
		tmpM := *(s.Memory)
		ret.Memory = &tmpM
	}
	return ret
}
func (s *SingleJobAllocationModel) Merge(other SingleJobAllocationModel) {
	if other.Cluster != "" {
		s.Cluster = other.Cluster
	}
	if other.Partition != "" {
		s.Partition =
			other.Partition
	}
	other.Cpu = CopyIntPointer(s.Cpu)
	other.Gpu = CopyIntPointer(s.Gpu)
	other.Memory = CopyIntPointer(s.Memory)
}

type JobAllocationModel struct {
	Default *SingleJobAllocationModel
	Diff    map[string]*SingleJobAllocationModel
}

func (j *JobAllocationModel) Get(path message.PathID) SingleJobAllocationModel {
	ret := j.Default.Copy()
	tmp, ok := j.Diff[path.Path()]
	if ok {
		ret.Merge(*tmp)
	}
	return ret
}

func (j *JobAllocationModel) Set(path message.PathID, changed SingleJobAllocationModel) {
	_, ok := j.Diff[path.Path()]
	if !ok {
		tmp := changed.Copy()
		j.Diff[path.Path()] = &tmp
	} else {
		j.Diff[path.Path()].Merge(changed)
	}
}

func NewSubmitModelFrom(allocation SingleJobAllocationModel) JobSubmitModel {
	ret := JobSubmitModel{}
	ret.RuntimeParams.Cluster = allocation.Cluster
	ret.RuntimeParams.Partition = allocation.Partition

	if allocation.Cpu != nil {
		ret.RuntimeParams.Cpu = *(allocation.Cpu)
	}
	if allocation.Gpu != nil {
		ret.RuntimeParams.Gpu = *(allocation.Gpu)
	}
	if allocation.Memory != nil {
		ret.RuntimeParams.Memory = *(allocation.Memory)
	}
	return ret
}
