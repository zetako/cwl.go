package runner

import (
	"errors"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/runner/message"
	"github.com/zetako/scontrol"
	"log"
	"time"
)

type WorkflowRunner struct {
	engine            *Engine
	process           *Process
	workflow          *cwl.Workflow
	runningCounter    int
	conditionChan     chan Condition
	neededConditions  []Condition
	steps             []StepRunner
	reachedConditions []Condition
	parameter         *cwl.Values
}

func (r *WorkflowRunner) MeetConditions(now []Condition) bool {
	for index := range r.neededConditions {
		if !r.neededConditions[index].Meet(now) {
			return false
		}
	}
	return true
}

// Run 执行工作流
//   - 该函数线程不安全，一个工作流实例仅可执行一个
func (r *WorkflowRunner) Run(channel chan<- Condition) (err error) {
	var (
		tmpCondition  Condition
		moreCondition bool
	)
	// 发送初始化信息
	var stepNames []string
	for _, step := range r.workflow.Steps {
		stepNames = append(stepNames, step.ID)
	}
	r.engine.SendMsg(message.Message{
		Class:     message.WorkflowMsg,
		Status:    message.StatusInit,
		ID:        r.process.PathID,
		TimeStamp: time.Now(),
		Content:   stepNames,
	})
	// 遍历一遍，全部尝试启动
	var tmpSteps []StepRunner
	for index := range r.steps {
		if r.steps[index].RunAtMeetConditions(r.reachedConditions, r.conditionChan, *r.parameter) {
			r.runningCounter++
		} else {
			tmpSteps = append(tmpSteps, r.steps[index])
		}
		// 读取flag内的最大并行任务数
		if r.engine.Flags.MaxParallelLimit > 0 && r.runningCounter > r.engine.Flags.MaxParallelLimit {
			break
		}
	}
	r.steps = tmpSteps
	for r.runningCounter > 0 {
		if r.engine.controller.Check() == scontrol.StatusStop {
			return fmt.Errorf("SignalAbort received")
		}
		tmpCondition = <-r.conditionChan // 接收到步骤完成的条件
		moreCondition = true
		if doneCond, ok := tmpCondition.(*StepDoneCondition); ok {
			var err error
			r.parameter, err = mergeStepOutputs(r.parameter, *doneCond)
			if err != nil {
				return err
			}
			r.runningCounter--
		}
		if errCond, ok := tmpCondition.(*StepErrorCondition); ok {
			return errCond.err
		}
		r.reachedConditions = append(r.reachedConditions, tmpCondition)

		for moreCondition {
			select {
			case tmpCondition = <-r.conditionChan:
				if doneCond, ok := tmpCondition.(*StepDoneCondition); ok {
					var err error
					r.parameter, err = mergeStepOutputs(r.parameter, *doneCond)
					if err != nil {
						return err
					}
					r.runningCounter--
				}
				if errCond, ok := tmpCondition.(*StepErrorCondition); ok {
					return errCond.err
				}
				r.reachedConditions = append(r.reachedConditions, tmpCondition)
				moreCondition = true
			default:
				moreCondition = false
			}
		}

		// 遍历一遍，全部尝试启动
		tmpSteps = []StepRunner{}
		for index := range r.steps {
			if r.steps[index].RunAtMeetConditions(r.reachedConditions, r.conditionChan, *r.parameter) {
				r.runningCounter++
			} else {
				tmpSteps = append(tmpSteps, r.steps[index])
			}
		}
		r.steps = tmpSteps
	}
	// 生成输出
	outputs := cwl.Values{}
	for _, output := range r.workflow.Outputs {
		if workflowOutput, ok := output.(*cwl.WorkflowOutputParameter); ok {
			if workflowOutput.PickValue != nil && *workflowOutput.PickValue != "" {
				// 有pickValue，先产生数组，然后pick
				var valueArr []cwl.Value
				// linkMerge
				value, err := linkMerge(workflowOutput.LinkMerge, workflowOutput.OutputSource, *r.parameter)
				if err != nil {
					return err
				}
				// pickValue
				if valueArr, ok = value.([]cwl.Value); ok {
					value, err = pickValue(valueArr, *workflowOutput.PickValue)
					if err != nil {
						return err
					}
				}
				outputs[workflowOutput.ID] = value
			} else {
				value, err := linkMerge(workflowOutput.LinkMerge, workflowOutput.OutputSource, *r.parameter)
				if err != nil {
					return err
				}
				outputs[workflowOutput.ID] = value
			}
		} else {
			return errors.New("输出不匹配")
		}
	}
	channel <- &WorkflowEndCondition{Out: outputs}
	return nil
}

func (r *WorkflowRunner) RunAtMeetConditions(now []Condition, channel chan<- Condition, parameter cwl.Values) (run bool) {
	if r.MeetConditions(now) {
		go func() {
			err := r.Run(channel)
			if err != nil {
				log.Println(err)
			}
		}()
		return true
	}
	return false
}

func (r *WorkflowRunner) SetInput(values cwl.Values) error {
	return fmt.Errorf("TODO ")
}

func (r *WorkflowRunner) GetPath() message.PathID {
	return r.process.PathID
}

func NewWorkflowRunner(e *Engine, wf *cwl.Workflow, p *Process, inputs *cwl.Values) (*WorkflowRunner, error) {
	r := &WorkflowRunner{
		engine:   e,
		process:  p,
		workflow: wf,
	}
	// 初始化需求的条件 TODO
	// 初始化每一步的执行器
	r.parameter = inputs
	if r.parameter == nil {
		r.parameter = &cwl.Values{}
	}
	r.steps = []StepRunner{}
	for _, step := range wf.Steps {
		tmpStep := step
		tmp, err := NewStepRunner(e, r, &tmpStep)
		if err != nil {
			return nil, err
		}
		r.steps = append(r.steps, tmp)
	}
	// 初始化输入条件
	r.reachedConditions = []Condition{}
	for _, in := range r.workflow.Inputs {
		wfIn, ok := in.(*cwl.WorkflowInputParameter)
		key := wfIn.ID
		value, ok := (*inputs)[key]
		if !ok {
			if wfIn.Type.IsNullable() {
				r.reachedConditions = append(r.reachedConditions, WorkflowInitCondition{
					key:   key,
					value: nil,
				})
			} else {
				return nil, fmt.Errorf("一个非空输入的值没有定义")
			}
		} else {
			r.reachedConditions = append(r.reachedConditions, WorkflowInitCondition{
				key:   key,
				value: value,
			})
		}
	}
	r.conditionChan = make(chan Condition, 10)
	r.runningCounter = 0
	return r, nil
}

// mergeStepOutputs 将 StepDoneCondition 的输出合并到指定数值列表中
// TODO 重构为 StepDoneCondition 的方法
func mergeStepOutputs(ori *cwl.Values, stepDone StepDoneCondition) (*cwl.Values, error) {
	// 0.预处理空值
	// 0.1 空输出不需要处理
	if stepDone.out == nil || *stepDone.out == nil { // 是空的，不需要输出
		return ori, nil
	}
	// 0.2 空输入需要初始化
	if ori == nil {
		ori = &cwl.Values{}
	}
	// 1. 处理每个数据
	for key, value := range *stepDone.out {
		// 合并值
		(*ori)[stepDone.step.ID+"/"+key] = stepDone.AddStepInfoFor(value)
	}
	return ori, nil
}

func (r *WorkflowRunner) SetRecoverFlag() {
	r.workflow.NeedRecovered = true
}
