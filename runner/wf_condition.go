package runner

import (
	"github.com/lijiang2014/cwl.go"
	"strings"
)

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
	// 暂时只进行最简单的判断，即：
	//   1. 暂时只判断i.input.Source[0]
	//   2. 如果它没有 "/"，说明是一个初始条件，匹配初始条件
	//   3. 如果有，匹配输出条件
	if len(i.input.Source) <= 0 {
		return false // 暂时无法判断
	}
	source := i.input.Source[0]
	if strings.Contains(source, "/") {
		// 需要匹配输出条件
		for _, cond := range condition {
			if outCond, ok := cond.(OutputParamCondition); ok {
				tmp := outCond.step.ID + outCond.output.ID
				if tmp == source {
					return true
				}
			}
		}
	} else {
		for _, cond := range condition {
			if initCond, ok := cond.(WorkflowInitCondition); ok {
				if initCond.key == source {
					return true
				}
			}
		}
	}
	return false
}

type OutputParamCondition struct {
	step   *cwl.WorkflowStep
	output *cwl.WorkflowStepOutput
}

func (o OutputParamCondition) Meet(condition []Condition) bool {
	// 完全一致时返回True
	for _, cond := range condition {
		if outCond, ok := cond.(OutputParamCondition); ok {
			if outCond.step.ID == o.step.ID && outCond.output.ID == o.step.ID {
				return true
			}
		}
	}
	return false
}

type StepDoneCondition struct {
	step *cwl.WorkflowStep
	out  *cwl.Values
}

func (s StepDoneCondition) Meet(condition []Condition) bool {
	return true
}

type StepErrorCondition struct {
	step *cwl.WorkflowStep
	err  error
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

type WorkflowInitCondition struct {
	key   string
	value cwl.Value
}

func (w WorkflowInitCondition) Meet(condition []Condition) bool {
	return true
}
