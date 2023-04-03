package runner

import (
	"errors"
	"github.com/lijiang2014/cwl.go"
	"log"
)

// StepRunner 对应CWL的一个任务，即
//   - CommandLineTool
//   - Expression
//   - Workflow
type StepRunner interface {
	MeetConditions(now []Condition) bool
	Run(chan<- Condition) error
	RunAtMeetConditions(now []Condition, channel chan<- Condition) (run bool)
}

// RegularRunner 常规Step的执行器
//   - 用来执行CommandLineTool和Expression
//   - 原理上，应该也可以作为Workflow的执行器
type RegularRunner struct {
	neededCondition    []Condition
	engine             *Engine
	step               *cwl.WorkflowStep
	process            *Process
	parameter          *cwl.Values
	useWorkflowDefault map[string]bool
}

func (r *RegularRunner) MeetConditions(now []Condition) bool {
	// 比较条件集
	// 老的不考虑默认值的版本
	//for _, need := range r.neededCondition {
	//	if !need.Meet(now) {
	//		return false
	//	}
	//}
	for _, need := range r.neededCondition {
		inputCond, ok := need.(InputParamCondition)
		if !ok {
			if !need.Meet(now) {
				return false
			}
		} else {
			meet, useDefault := inputCond.MeetOrDefault(now)
			if !meet {
				return false
			}
			if useDefault {
				r.useWorkflowDefault[inputCond.input.ID] = true
			}
		}
	}
	return true
}

func (r *RegularRunner) Run(conditions chan<- Condition) (err error) {
	// 0. 如果需要Scatter，任务交由RunScatter
	if r.step.Scatter != nil && len(r.step.Scatter) > 0 {
		return r.RunScatter(conditions)
	}
	// 1. 先创建对应的 Process
	// 已经在创建 RegularRunner 时处理，不再需要
	//r.process, err = r.engine.GenerateSubProcess(r.step)
	//if err != nil {
	//	conditions <- &StepErrorCondition{
	//		step: r.step,
	//		err:  err,
	//	}
	//	return err
	//}
	// 2. 处理Input
	// TODO 必须根据每一步的需要单独绑定
	//r.process.inputs = r.parameter
	r.process.inputs = &cwl.Values{}
	for _, in := range r.step.In {
		if _, ok := r.useWorkflowDefault[in.ID]; ok {
			(*r.process.inputs)[in.ID] = in.Default
		} else if len(in.Source) == 1 {
			(*r.process.inputs)[in.ID] = (*r.parameter)[in.Source[0]]
		} else {
			switch in.LinkMerge {
			case "merge_flattened":
				var tmpInput []cwl.Value
				for _, src := range in.Source {
					tmp, ok := (*r.parameter)[src]
					if !ok {
						return errors.New("没有匹配的输入")
					}
					if tmpList, ok := tmp.([]cwl.Value); ok {
						tmpInput = append(tmpInput, tmpList...)
					} else { // 也有可能是单个元素
						tmpInput = append(tmpInput, tmp)
					}
				}
				(*r.process.inputs)[in.ID] = tmpInput
				break
			case "merge_nested":
				fallthrough
			default:
				(*r.process.inputs)[in.ID] = (*r.parameter)[in.Source[0]]
				break
			}
		}
	}

	// 3. 然后使用 Engine.RunProcess()
	outs, err := r.engine.RunProcess(r.process)
	if err != nil {
		conditions <- &StepErrorCondition{
			step: r.step,
			err:  err,
		}
		return err
	}
	// 4. 最后，根据Step的输出，释放Condition
	for _, output := range r.step.Out {
		conditions <- &OutputParamCondition{
			step:   r.step,
			output: &output,
		}
	}
	conditions <- &StepDoneCondition{
		step:    r.step,
		out:     &outs,
		runtime: r.process.runtime,
	}
	// 4. 返回
	return nil
}

func (r *RegularRunner) RunAtMeetConditions(now []Condition, channel chan<- Condition) (run bool) {
	if r.MeetConditions(now) {
		log.Println("Run Step: ", r.step.ID)
		go func() {
			err := r.Run(channel) // 暂时不管错误，最好把他输出到Log那里
			if err != nil {
				log.Println(err)
			}
		}()
		return true
	}
	return false
}

func NewStepRunner(e *Engine, step *cwl.WorkflowStep, param *cwl.Values) (StepRunner, error) {
	var (
		err error
	)
	// 目前返回的均为RegularRunner，未来可能考虑返回WorkflowRunner
	ret := RegularRunner{
		neededCondition:    []Condition{},
		engine:             e,
		step:               step,
		process:            nil,
		parameter:          param,
		useWorkflowDefault: map[string]bool{},
	}
	// 生成条件集
	for _, input := range step.In {
		ret.neededCondition = append(ret.neededCondition, InputParamCondition{
			step:  ret.step,
			input: input,
		})
	}
	// 创建process
	ret.process, err = ret.engine.GenerateSubProcess(ret.step)
	if err != nil {
		return nil, err
	}
	// 返回
	return &ret, nil
}
