package runner

import (
	"errors"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"log"
	"time"
)

type WorkflowRunner struct {
	engine            *Engine
	workflow          *cwl.Workflow
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
func (r *WorkflowRunner) Run(channel chan<- Condition) error {
	var (
		tmpCondition     Condition
		moreCondition    bool
		runningCounter   int = 0
		conditionChannel chan Condition
		ctrlSignal       Signal
	)
	// 发送初始化信息
	var stepNames []string
	for _, step := range r.workflow.Steps {
		stepNames = append(stepNames, step.ID)
	}
	r.engine.SendMsg(Message{
		Class:     WorkflowMsg,
		Status:    StatusInit,
		TimeStamp: time.Now(),
		Content:   stepNames,
	})
	conditionChannel = make(chan Condition)
	// 遍历一遍，全部尝试启动
	var tmpSteps []StepRunner
	for index := range r.steps {
		if r.steps[index].RunAtMeetConditions(r.reachedConditions, conditionChannel) {
			runningCounter++
		} else {
			tmpSteps = append(tmpSteps, r.steps[index])
		}
		// 读取flag内的最大并行任务数
		if r.engine.Flags.MaxParallelLimit > 0 && runningCounter > r.engine.Flags.MaxParallelLimit {
			break
		}
	}
	r.steps = tmpSteps
	for runningCounter > 0 {
		select {
		case ctrlSignal = <-r.engine.SignalChannel: // 接收到了控制信号
			r.SendCtrlSignal(ctrlSignal)
			switch ctrlSignal {
			case SignalAbort:
				// 直接中断执行
				return fmt.Errorf("SignalAbort received")
			case SignalPause:
			pausing: // 在这里阻塞，除非收到了恢复或中止信号
				ctrlSignal = <-r.engine.SignalChannel
				switch ctrlSignal {
				case SignalResume:
					// 可以继续
					break
				case SignalAbort:
					// 直接中断执行
					return fmt.Errorf("SignalAbort received")
				default:
					goto pausing
				}
			case SignalResume:
				fallthrough
			default:
				// DO NOTHING
				break
			}
		case tmpCondition = <-conditionChannel: // 接收到步骤完成的条件
			moreCondition = true
			if doneCond, ok := tmpCondition.(*StepDoneCondition); ok {
				var err error
				r.parameter, err = mergeStepOutputs(r.parameter, *doneCond)
				if err != nil {
					return err
				}
				runningCounter--
			}
			if errCond, ok := tmpCondition.(*StepErrorCondition); ok {
				return errCond.err
			}
			r.reachedConditions = append(r.reachedConditions, tmpCondition)

			// 没有default，上述两个任意命中才需要继续
			//default:
			//	// DO NOTHING
			//	break
		}

		for moreCondition {
			select {
			case tmpCondition = <-conditionChannel:
				if doneCond, ok := tmpCondition.(*StepDoneCondition); ok {
					var err error
					r.parameter, err = mergeStepOutputs(r.parameter, *doneCond)
					if err != nil {
						return err
					}
					runningCounter--
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
			if r.steps[index].RunAtMeetConditions(r.reachedConditions, conditionChannel) {
				runningCounter++
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

func (r *WorkflowRunner) RunAtMeetConditions(now []Condition, channel chan<- Condition) (run bool) {
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

func (r *WorkflowRunner) SendCtrlSignal(signal Signal) {
	for _, step := range r.steps {
		step.SendCtrlSignal(signal)
	}
}

func NewWorkflowRunner(e *Engine, wf *cwl.Workflow, inputs *cwl.Values) (*WorkflowRunner, error) {
	r := &WorkflowRunner{
		engine:   e,
		workflow: wf,
	}
	// 初始化需求的条件 TODO
	// 初始化每一步的执行器
	r.parameter = inputs
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
