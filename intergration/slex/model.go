package slex

import (
	"github.com/lijiang2014/cwl.go/message"
	"starlight/common/model"
)

type JobSubmitModel struct {
	RuntimeParams struct {
		// Basic Params
		JobName   string            `json:"jobname"`
		Cluster   string            `json:"cluster"`
		Partition string            `json:"partition"`
		Env       map[string]string `json:"env,omitempty"`
		Endpoints []model.Proxy     `json:"endpoints,omitempty"`
		Stdin     string            `json:"stdin,omitempty"`
		Stdout    string            `json:"stdout,omitempty"`
		Stderr    string            `json:"stderr,omitempty"`
		// Executor infos
		AppName      string `json:"app_name,omitempty"`
		WorkflowUUID string `json:"workflow_uuid,omitempty"`

		// Direct Task specified
		WorkDir model.Volume `json:"workDir,omitempty"`
		Cmd     []string     `json:"cmd"`
		// Container Task specified
		Image   string         `json:"image,omitempty"`
		Volumes []model.Volume `json:"volumes,omitempty"`
		//Kind   string         `json:"kind,omitempty"` // Kind not needed, for workflow, always job

		// Runtime Requirements
		Cpu    int `json:"cpu,omitempty"`
		Gpu    int `json:"gpu,omitempty"`
		Memory int `json:"memory,omitempty"`
		Node   int `json:"node,omitempty"`
	} `json:"runtime_params"`
}

type SingleJobAllocationModel struct {
	Cluster   string       `json:"cluster" yaml:"cluster"`
	Partition string       `json:"partition" yaml:"partition"`
	Cpu       *int         `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Gpu       *int         `json:"gpu,omitempty" yaml:"gpu,omitempty"`
	Memory    *int         `json:"memory,omitempty" yaml:"memory,omitempty"`
	Node      *int         `json:"node,omitempty" yaml:"node,omitempty"`
	WorkDir   model.Volume `json:"workDir,omitempty" yaml:"workDir,omitempty"`
}

func (s *SingleJobAllocationModel) Copy() SingleJobAllocationModel {
	ret := SingleJobAllocationModel{
		Cluster:   s.Cluster,
		Partition: s.Partition,
		WorkDir:   s.WorkDir,
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
	if s.Node != nil {
		tmpN := *(s.Node)
		ret.Node = &tmpN
	}
	return ret
}
func (s *SingleJobAllocationModel) Merge(other SingleJobAllocationModel) {
	if other.Cluster != "" {
		s.Cluster = other.Cluster
	}
	if other.Partition != "" {
		s.Partition = other.Partition
	}
	if other.WorkDir.HostPath != "" {
		s.WorkDir = other.WorkDir
	}
	other.Cpu = CopyIntPointer(s.Cpu)
	other.Gpu = CopyIntPointer(s.Gpu)
	other.Memory = CopyIntPointer(s.Memory)
	other.Node = CopyIntPointer(s.Node)
}

type JobAllocationModel struct {
	Default *SingleJobAllocationModel            `json:"default" yaml:"default"`
	Diff    map[string]*SingleJobAllocationModel `json:"diff" yaml:"diff"`
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
	if j.Diff == nil {
		j.Diff = map[string]*SingleJobAllocationModel{}
	}
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
	ret.RuntimeParams.WorkDir = allocation.WorkDir

	if allocation.Cpu != nil {
		ret.RuntimeParams.Cpu = *(allocation.Cpu)
	}
	if allocation.Gpu != nil {
		ret.RuntimeParams.Gpu = *(allocation.Gpu)
	}
	if allocation.Memory != nil {
		ret.RuntimeParams.Memory = *(allocation.Memory)
	}
	if allocation.Node != nil {
		ret.RuntimeParams.Node = *(allocation.Node)
	}
	return ret
}
