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

	// 尝试先引入所有文档
	if err = e.tryImportRun(wf, p.root.Graph, 0); err != nil {
		return nil, err
	}

	wfRunner, err := NewWorkflowRunner(e, wf, p, p.inputs)
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
