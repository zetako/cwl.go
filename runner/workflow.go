package runner

import (
	"fmt"

	"github.com/lijiang2014/cwl.go"
)

//
func (p *Process) RunWorflow() (cwl.Values, error) {
	wf, ok := p.root.Process.(*cwl.Workflow)
	if !ok {
		return nil, fmt.Errorf("not Workflow")
	}
	// for _, inb := range wf.Inputs {
	// 	// var binding *cwl.CommandLineBinding
	// 	// in := inb.(*cwl.WorkflowInputParameter)
	// 	// val := (*p.inputs)[in.ID]
	// 	// k := sortKey{0}
	// 	// if in.InputBinding != nil {
	// 	// 	binding = &cwl.CommandLineBinding{InputBinding: *in.InputBinding}
	// 	// }
	// 	// b, err := p.bindInput(in.ID, in.Type, binding, in.SecondaryFiles, val, k)
	// 	// if err != nil {
	// 	// 	return nil, p.errorf("binding input %q: %s", in.ID, err)
	// 	// }

	// 	// if b == nil {
	// 	// 	return nil, p.errorf("no binding found for input: %s", in.ID)
	// 	// }
	// 	// p.bindings = append(p.bindings, b...)
	// 	_ = p
	// }

	// TODO  exec workflow
	_ = wf

	var out interface{}
	var err error
	// Convert out into values
	valMap, ok := out.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("output format err")
	}
	values := cwl.Values{}
	for key, v := range valMap {
		newv, err := cwl.ConvertToValue(v)
		if err != nil {
			return nil, err
		}
		values[key] = newv
	}
	return values, err
}
