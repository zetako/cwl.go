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
	// Meet 在参数指定的条件列表下，本条件已经满足
	Meet(condition []Condition) bool
}

// InputParamCondition 指定步骤需要的输入参数
type InputParamCondition struct {
	step  *cwl.WorkflowStep
	input cwl.WorkflowStepInput
}

// Meet 在参数指定的条件列表下，本条件已经满足
//   - 有对应的初始条件 WorkflowInitCondition 或者之前步骤的输出条件 OutputParamCondition
//   - 如果有Default, 也是满足的
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
			if outCond, ok := cond.(*OutputParamCondition); ok {
				tmp := outCond.step.ID + "/" + outCond.output.ID
				if tmp == source {
					return true
				}
			}
		}
	} else {
		// 需匹配初始条件
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

func (i InputParamCondition) MeetOrDefault(condition []Condition) (meet, useDefault bool) {
	if i.Meet(condition) {
		return true, false
	}
	if i.input.Default != nil {
		return true, true
	}
	return false, false
}

// OutputParamCondition 完成步骤后的输出
type OutputParamCondition struct {
	step   *cwl.WorkflowStep
	output *cwl.WorkflowStepOutput
}

// Meet 在参数指定的条件列表下，本条件已经满足
//   - 具有相同步骤的相同输出时满足
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

// StepDoneCondition 步骤正常结束的条件，包含输出
type StepDoneCondition struct {
	step    *cwl.WorkflowStep
	out     *cwl.Values
	runtime Runtime
}

// Meet 无意义的判断，始终满足
func (s StepDoneCondition) Meet(condition []Condition) bool {
	return true
}

// StepErrorCondition 步骤错误结束的条件，包含错误
type StepErrorCondition struct {
	step *cwl.WorkflowStep
	err  error
}

// Meet 无意义的判断，始终满足
func (s StepErrorCondition) Meet(condition []Condition) bool {
	return true
}

// WorkflowEndCondition 工作流正常结束的条件，包含输出
type WorkflowEndCondition struct {
	Out cwl.Values
}

// Meet 无意义的判断，始终满足
func (w WorkflowEndCondition) Meet(condition []Condition) bool {
	return true
}

// WorkflowInitCondition 工作流初始化时的输入
type WorkflowInitCondition struct {
	key   string
	value cwl.Value
}

// Meet 无意义的判断，始终满足
func (w WorkflowInitCondition) Meet(condition []Condition) bool {
	return true
}
