package runner

import "github.com/lijiang2014/cwl.go"

// Condition 可以是任何一种条件，例如：
//   - 任务完成
//   - 文件需求
//   - 资源需求
//   - ……
type Condition interface {
	Meet(condition []Condition) bool
}

type InputParamCondition struct {
	step  *cwl.WorkflowStep
	input *cwl.WorkflowStepInput
}

func (i InputParamCondition) Meet(condition []Condition) bool {
	//TODO implement me
	panic("implement me")
	// 有满足条件的output时返回True
}

type OutputParamCondition struct {
	step   *cwl.WorkflowStep
	output *cwl.WorkflowStepOutput
}

func (o OutputParamCondition) Meet(condition []Condition) bool {
	//TODO implement me
	panic("implement me")
	// 完全一致时返回True
}

type StepDoneCondition struct {
	step *cwl.WorkflowStep
}

func (s StepDoneCondition) Meet(condition []Condition) bool {
	return true
}

type StepErrorCondition struct {
	step *cwl.WorkflowStep
}

func (s StepErrorCondition) Meet(condition []Condition) bool {
	return true
}

type WorkflowEndCondition struct {
	Out map[string]interface{}
}

func (w WorkflowEndCondition) Meet(condition []Condition) bool {
	return true
}
