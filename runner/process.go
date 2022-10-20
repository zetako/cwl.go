package runner

import (
	"encoding/json"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"log"
	"sort"
	"strings"
)

type Process struct {
	root           *cwl.Root
	tool	         *cwl.CommandLineTool
	inputs         *cwl.Values
	runtime        Runtime
	fs             Filesystem
	bindings       []*Binding
	expressionLibs []string
	env            map[string]string
	filesToCreate  []cwl.FileDir
	shell          bool
	resources      ResourcesLimites
	stdout         string
	stderr         string
	*Log
	*jsvm
}

func (p *Process) Root() *cwl.Root {
	return p.root
}


// Binding binds an input type description (string, array, record, etc)
// to a concrete input value. this information is used while building
// command line args.
type Binding struct {
	clb *cwl.CommandLineBinding
	// the bound type (resolved by matching the input value to one of many allowed types)
	// can be nil, which means no matching type could be determined.
	Type cwl.SaladType
	// the value from the input object
	Value cwl.Value
	// used to determine the ordering of command line flags.
	// http://www.commonwl.org/v1.0/CommandLineTool.html#Input_binding
	sortKey sortKey
	nested  []*Binding
	name    string
}

// 多层的排序 key
type sortKey []interface{}

func setDefault(values *cwl.Values, inputs cwl.Inputs) {
	for _, in := range inputs {
		any, ok := (*values)[in.GetInputParameter().ID]
		if (!ok || any == nil )&& in.GetInputParameter().Default != nil {
			(*values)[in.GetInputParameter().ID] = in.GetInputParameter().Default
		}
	}
}

func (process *Process) Command() ([]string, error) {
	// Copy "Tool.Inputs" bindings
	args := make([]*Binding, 0, len(process.bindings))
	for _, b := range process.bindings {
		if b.clb != nil {
			args = append(args, b)
		}
	}
	tool := process.root.Process.(*cwl.CommandLineTool)
	// Add "Tool.Arguments"
	for i, arg := range  tool.Arguments {
		if arg.Binding == nil && arg.Exp != "" {
			expStr := string(arg.Exp)
			expResult , expErr :=  process.Eval(arg.Exp, nil)
			if expErr == nil {
				expStr = fmt.Sprint(expResult)
			}
			args = append(args, &Binding{
				arg.Binding,  cwl.NewType(argType), expStr, sortKey{0}, nil, "",
			})
			continue
		} else if arg.Binding == nil {
			return nil, fmt.Errorf("empty argument")
		}
		if arg.Binding != nil && arg.Binding.ValueFrom == ""  {
			return nil, fmt.Errorf("valueFrom is required but missing for argument %d", i)
		}
		args = append(args, &Binding{
			arg.Binding, cwl.NewType(argType), nil, sortKey{arg.Binding.Position}, nil, "",
		})
	}
	//
	// Evaluate "valueFrom" expression.
	for _, b := range args {
		if b.clb != nil && b.clb.ValueFrom != ""  {
			val, err := process.eval(b.clb.ValueFrom, b.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to eval argument value: %s", err)
			}
			b.Value = val
		}
	}
	
	sort.Stable(bySortKey(args))
	//debug(args)
	//
	//// Now collect the input bindings into command line arguments
	cmd := append([]string{}, tool.BaseCommands...)
	for _, b := range args {
		cmd = append(cmd, bindArgs(b)...)
	}
	//
	if tool.RequiresShellCommand() {
		cmd = []string{"/bin/sh", "-c", strings.Join(cmd, " ")}
	}

	//debug("COMMAND", cmd)
	return cmd, nil
}

func (p *Process) loadRuntime() {
	p.vm.Set("runtime", p.runtime)
}

//
func (p *Process) RunExpression() (cwl.Values, error) {
	tool, ok := p.root.Process.(*cwl.ExpressionTool)
	if !ok {
		return nil, fmt.Errorf("not ExpressionTool")
	}
	for _, inb := range tool.Inputs {
		var binding *cwl.CommandLineBinding
		in := inb.(*cwl.WorkflowInputParameter)
		val := (*p.inputs)[in.ID]
		k := sortKey{0}
		if in.InputBinding != nil {
			binding = &cwl.CommandLineBinding{ InputBinding: *in.InputBinding }
		}
		b, err := p.bindInput(in.ID, in.Type, binding, in.SecondaryFiles, val, k)
		if err != nil {
			return nil,  p.errorf("binding input %q: %s", in.ID, err)
		}
		if b == nil {
			return nil,  p.errorf("no binding found for input: %s", in.ID)
		}
		p.bindings = append(p.bindings, b...)
	}
	if err := p.initJVM(); err != nil {
		return nil, err
	}
	
	out, err := p.jsvm.Eval(tool.Expression,nil)
	log.Printf("out %#v",out)
	// TODO Convert out into values
	valMap,ok := out.(map[string]interface{})
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


func toJSONMap(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	// Need to convert Go variable naming to JSON. Easiest way to to marshal to JSON,
	// then unmarshal into a map.
	j, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal(j, &data)
	if err != nil {
		return nil, fmt.Errorf(`marshaling data for JS evaluation %s`, err)
	}
	return data, nil
}

func (process *Process) eval(x cwl.Expression, self interface{}) (interface{}, error) {
	return process.jsvm.Eval(x, self)
}

func (process *Process) Env() map[string]string {
	env := map[string]string{}
	for k, v := range process.env {
		env[k] = v
	}
	return env
}

func (process *Process) loadReqs() error {
	tool := process.root.Process.(*cwl.CommandLineTool)
	if req := tool.RequiresInlineJavascript(); req != nil {
		// TODO Add Libs
	 if err := process.initJVM(); err != nil {
	   return err
	 }
	 for _, lib := range req.ExpressionLib {
	 	//out, err := process.eval(cwl.Expression(lib), nil )
	 	v, err := process.jsvm.vm.Run(lib)
	 	if err != nil {
	 		log.Println(v, err)
		}
	 }
	}
	if req := tool.RequiresEnvVar(); req != nil {
		// TODO env
	}
	if req := tool.RequiresResource(); req != nil {
		process.loadRuntime()
	}
	if req := tool.RequiresSchemaDef(); req != nil {
		// TODO init schemaDef
	}
	if req := tool.RequiresInitialWorkDir(); req != nil {
		// TODO env
	}
	return nil
}


func (process *Process) ResourcesLimites() (*ResourcesLimites , error){
	limits := GetDefaultResourcesLimits()
	if req := process.tool.RequiresResource(); req != nil {
		//CoresMin        LongFloatExpression `json:"coresMin,omitempty"`
		//CoresMax        LongFloatExpression `json:"coresMax,omitempty"`
		//RamMin          LongFloatExpression `json:"ramMin,omitempty"` // Minimum reserved RAM in mebibytes (2**20) (default is 256)
		//RamMax          LongFloatExpression `json:"ramMax,omitempty"`
		//TmpdirMin       LongFloatExpression `json:"tmpdirMin,omitempty"` // Minimum reserved filesystem based storage for the designated temporary directory, in mebibytes (2**20) (default is 1024)
		//TmpdirMax       LongFloatExpression `json:"tmpdirMax,omitempty"`
		//OutdirMin       LongFloatExpression `json:"outdirMin,omitempty"` // Minimum reserved filesystem based storage for the designated output directory, in mebibytes (2**20) (default is 1024)
		//OutdirMax       LongFloatExpression `json:"outdirMax,omitempty"`
		
		if !req.CoresMin.IsNull() {
			err := req.CoresMin.Resolve(process.jsvm, nil)
			if err != nil {
				return nil, fmt.Errorf("ResourcesLimites CoresMin Expression Resolve Error %s", err)
			}
			limits.CoresMin = req.CoresMin.MustInt64()
		}
		// TODO Set more Resources
	}
	return &limits, nil
}