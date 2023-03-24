package runner

import (
	"errors"
	"github.com/lijiang2014/cwl.go"
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
	neededCondition []Condition
	engine          *Engine
	step            *cwl.WorkflowStep
	process         *Process
}

func (r *RegularRunner) MeetConditions(now []Condition) bool {
	// 生成条件集
	if r.neededCondition == nil || len(r.neededCondition) == 0 {
		r.neededCondition = []Condition{}
		for _, input := range r.step.In {
			r.neededCondition = append(r.neededCondition, InputParamCondition{
				step:  r.step,
				input: &input,
			})
		}
	}
	// 比较条件集
	for _, need := range r.neededCondition {
		if !need.Meet(now) {
			return false
		}
	}
	return true
}

func (r *RegularRunner) Run(conditions chan<- Condition) (err error) {
	// 1. 先创建对应的 Process
	r.process, err = r.engine.GenerateSubProcess(r.step)
	if err != nil {
		conditions <- &StepErrorCondition{step: r.step}
		return err
	}
	// 2. 处理Input

	// TODO

	// 3. 然后使用 Engine.RunProcess()
	_, err = r.engine.RunProcess(r.process)
	if err != nil {
		conditions <- &StepErrorCondition{step: r.step}
		return err
	}
	// 4. 最后，根据Step的输出，释放Condition
	for _, output := range r.step.Out {
		conditions <- &OutputParamCondition{
			step:   r.step,
			output: &output,
		}
	}
	conditions <- &StepDoneCondition{step: r.step}
	// 4. 返回
	return nil
}

func (r *RegularRunner) RunAtMeetConditions(now []Condition, channel chan<- Condition) (run bool) {
	if r.MeetConditions(now) {
		go func() {
			_ = r.Run(channel) // 暂时不管错误，最好把他输出到Log那里
		}()
		return true
	}
	return false
}

func NewStepRunner(e *Engine, step *cwl.WorkflowStep) (StepRunner, error) {
	// TODO
	return nil, errors.New("NOT IMPLEMENTED")
}
