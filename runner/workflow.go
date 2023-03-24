package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
)

func (p *Process) RunWorkflow(e *Engine) (cwl.Values, error) {
	var (
		err        error
		outChannel chan Condition
		outValMap  map[string]interface{}
	)
	wf, ok := p.root.Process.(*cwl.Workflow)
	if !ok {
		return nil, fmt.Errorf("not Workflow")
	}

	wfRunner, err := NewWorkflowRunner(e, wf)
	if err != nil {
		return nil, err
	}
	err = wfRunner.Run(outChannel)
	if err != nil {
		return nil, err
	}
	select {
	case tmp := <-outChannel:
		if tmpWf, ok := tmp.(WorkflowEndCondition); ok {
			outValMap = tmpWf.Out
		} else {
			outValMap = nil
		}
	default:
		outValMap = nil
	}

	// Convert out into values
	if !ok {
		return nil, fmt.Errorf("output format err")
	}
	values := cwl.Values{}
	for key, v := range outValMap {
		newv, err := cwl.ConvertToValue(v)
		if err != nil {
			return nil, err
		}
		values[key] = newv
	}
	return values, err
}
