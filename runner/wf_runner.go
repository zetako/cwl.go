package runner

import (
	"github.com/lijiang2014/cwl.go"
	"log"
	"path"
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
	)
	conditionChannel = make(chan Condition)
	// 遍历一遍，全部尝试启动
	var tmpSteps []StepRunner
	for index := range r.steps {
		if r.steps[index].RunAtMeetConditions(r.reachedConditions, conditionChannel) {
			runningCounter++
		} else {
			tmpSteps = append(tmpSteps, r.steps[index])
		}
	}
	r.steps = tmpSteps
	for {
		// 接收新完成的条件
		tmpCondition = <-conditionChannel
		moreCondition = true
		if doneCond, ok := tmpCondition.(*StepDoneCondition); ok {
			mergeStepOutputs(r.parameter, *doneCond)
			runningCounter--
		}
		if errCond, ok := tmpCondition.(*StepErrorCondition); ok {
			return errCond.err
		}
		r.reachedConditions = append(r.reachedConditions, tmpCondition)
		for moreCondition {
			select {
			case tmpCondition = <-conditionChannel:
				if doneCond, ok := tmpCondition.(*StepDoneCondition); ok {
					mergeStepOutputs(r.parameter, *doneCond)
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

		// 检查是否还有任务运行
		if runningCounter <= 0 {
			break
		}
	}
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
		tmp, err := NewStepRunner(e, &tmpStep, r.parameter)
		if err != nil {
			return nil, err
		}
		r.steps = append(r.steps, tmp)
	}
	// 初始化输入条件
	r.reachedConditions = []Condition{}
	for key, value := range *inputs {
		r.reachedConditions = append(r.reachedConditions, WorkflowInitCondition{
			key:   key,
			value: value,
		})
	}
	return r, nil
}

func mergeCwlValues(base *cwl.Values, others ...*cwl.Values) *cwl.Values {
	if base == nil {
		base = &cwl.Values{}
	}
	for _, other := range others {
		for key, value := range *other {
			(*base)[key] = value
		}
	}
	return base
}

func mergeStepOutputs(parameter *cwl.Values, stepDone StepDoneCondition) *cwl.Values {
	if parameter == nil {
		parameter = &cwl.Values{}
	}

	for key, value := range *stepDone.out {
		var tmpValue cwl.Value
		if file, ok := value.(cwl.File); ok {
			if !path.IsAbs(file.Location) {
				file.Location = path.Join(stepDone.runtime.RootHost, file.Location)
			}
			tmpValue = file
		} else {
			tmpValue = value
		}

		(*parameter)[stepDone.step.ID+"/"+key] = tmpValue
	}
	return parameter
}
