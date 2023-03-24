package runner

import (
	"errors"
	"github.com/lijiang2014/cwl.go"
)

type WorkflowRunner struct {
	engine            *Engine
	workflow          *cwl.Workflow
	neededConditions  []Condition
	steps             []StepRunner
	reachedConditions []Condition
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
	for {
		// 遍历一遍，全部尝试启动
		for index := range r.steps {
			if r.steps[index].RunAtMeetConditions(r.reachedConditions, conditionChannel) {
				runningCounter++
				r.steps = append(r.steps[:index], r.steps[index+1:]...)
			}
		}

		// 接收新完成的条件
		tmpCondition = <-conditionChannel
		moreCondition = true
		if _, ok := tmpCondition.(StepDoneCondition); ok {
			runningCounter--
		}
		r.reachedConditions = append(r.reachedConditions, tmpCondition)
		for moreCondition {
			select {
			case tmpCondition = <-conditionChannel:
				if _, ok := tmpCondition.(StepDoneCondition); ok {
					runningCounter--
				}
				r.reachedConditions = append(r.reachedConditions, tmpCondition)
				moreCondition = true
			default:
				moreCondition = false
			}
		}

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
			_ = r.Run(channel)
		}()
		return true
	}
	return false
}

func NewWorkflowRunner(e *Engine, wf *cwl.Workflow) (*WorkflowRunner, error) {
	r := &WorkflowRunner{
		engine:   e,
		workflow: wf,
	}
	// 初始化需求的条件 TODO
	// 初始化每一步的执行器
	r.steps = []StepRunner{}
	for _, step := range wf.Steps {
		tmp, err := NewStepRunner(e, &step)
		if err != nil {
			return nil, err
		}
		r.steps = append(r.steps, tmp)
	}
	// 初始化当前条件 TODO
	return nil, errors.New("NOT IMPLEMENTED")
}
