package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	message2 "github.com/lijiang2014/cwl.go/message"
	"log"
	"time"
)

func (p *Process) RunWorkflow(e *Engine) (cwl.Values, error) {
	var (
		err        error
		outChannel chan Condition
		values     cwl.Values
	)
	outChannel = make(chan Condition, 1)
	wf, ok := p.root.Process.(*cwl.Workflow)
	if !ok {
		return nil, fmt.Errorf("not Workflow")
	}

	// 尝试先引入所有文档
	if err = e.tryImportRun(wf, p.root.Graph, 0); err != nil {
		return nil, err
	}

	wfRunner, err := NewWorkflowRunner(e, wf, p, p.inputs)
	if err != nil {
		return nil, err
	}
	err = wfRunner.tryRecover()
	if err != nil {
		return nil, err
	}
	err = wfRunner.Run(outChannel)
	if err != nil {
		return nil, err
	}
	select {
	case tmp := <-outChannel:
		if tmpWf, ok := tmp.(*WorkflowEndCondition); ok {
			values = tmpWf.Out
		} else {
			values = nil
		}
	default:
		values = nil
	}
	for _, out := range wf.Outputs {
		wfOut := out.(*cwl.WorkflowOutputParameter)
		p.fs = p.outputFS
		v, err := p.bindOutput(p.outputFS, wfOut.Type, nil, wfOut.SecondaryFiles, values[wfOut.ID])
		if err != nil {
			return nil, fmt.Errorf(`failed to bind value for "%s": %s`, wfOut.ID, err)
		}
		values[wfOut.ID] = v
	}
	return values, err
}

func (r *WorkflowRunner) tryRecover() error {
	// 没有标记的工作流不需要恢复
	if !r.workflow.NeedRecovered {
		return nil
	}
	if r.engine.ImportedStatus == nil {
		return fmt.Errorf("cannnot recovered without imported status")
	}

	// 若需要恢复
	// 先获取3种Step
	var (
		allSteps     *message2.StepStatusArray
		finishSteps  []*message2.StepStatus
		runningSteps []*message2.StepStatus
		waitingSteps []*message2.StepStatus
		tmpRunners   []StepRunner
		tmpCounter   int
	)
	allSteps = r.engine.ImportedStatus.GetTier(r.process.PathID)
	allSteps.Foreach(func(status *message2.StepStatus) {
		switch status.Status {
		case message2.StatusFinish:
			finishSteps = append(finishSteps, status)
		case message2.StatusStart:
			runningSteps = append(runningSteps, status)
		default:
			waitingSteps = append(waitingSteps, status)
		}
	})
	// 已完成的保存值
	if r.parameter == nil {
		r.parameter = &cwl.Values{}
	}
	for _, s := range finishSteps {
		// 记录结果
		for k, v := range s.Output {
			(*r.parameter)[s.ID.ID()+"/"+k] = v
		}
		// 找到对应的步骤
		var myStep cwl.WorkflowStep
		for _, step := range r.workflow.Steps {
			if step.ID == s.ID.ID() {
				myStep = step
			}
		}
		// 设置结果条件
		for _, out := range myStep.Out {
			r.reachedConditions = append(r.reachedConditions, &OutputParamCondition{
				step:   &myStep,
				output: &out,
			})
		}
		// 设置完成条件
		r.reachedConditions = append(r.reachedConditions, &StepDoneCondition{
			step: &myStep,
			out:  &s.Output,
		})
		// 发送步骤完成信号
		r.engine.SendMsg(message2.Message{
			Class:     message2.StepMsg,
			Status:    message2.StatusFinish,
			TimeStamp: time.Now(),
			ID:        nil,
			Index:     0,
			Content:   s.Output,
		})
	}
	// 进行中的调用协程启动
	for _, s := range runningSteps {
		// 找到对应的步骤
		var myStep cwl.WorkflowStep
		for _, step := range r.workflow.Steps {
			if step.ID == s.ID.ID() {
				myStep = step
			}
		}
		// 产生RegularRunner
		tmpRunner, err := NewStepRunner(r.engine, r, &myStep)
		if err != nil {
			return err
		}
		// 设置Input
		err = tmpRunner.SetInput(*r.parameter)
		if err != nil {
			return err
		}
		// 设置recover标志位
		tmpRunner.SetRecoverFlag()
		// 协程调用Run
		go func() {
			err = tmpRunner.Run(r.conditionChan)
			if err != nil {
				log.Println(err)
			}
		}()
		// 计数++
		tmpCounter++
	}
	r.runningCounter = tmpCounter
	// 否则加入等待列表
	for _, s := range waitingSteps {
		// 找到对应的步骤
		var myStep cwl.WorkflowStep
		for _, step := range r.workflow.Steps {
			if step.ID == s.ID.ID() {
				myStep = step
			}
		}
		// 产生RegularRunner
		tmpRunner, err := NewStepRunner(r.engine, r, &myStep)
		if err != nil {
			return err
		}
		// 加入列表
		tmpRunners = append(tmpRunners, tmpRunner)
	}
	r.steps = tmpRunners

	return nil
}
