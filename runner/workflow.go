package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
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

	wfRunner, err := NewWorkflowRunner(e, wf, p.inputs)
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
	// TODO 绑定输出 || 验证输出
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
