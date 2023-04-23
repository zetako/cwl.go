package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/lijiang2014/cwl.go"
)

type Process struct {
	root              *cwl.Root
	tool              *cwl.CommandLineTool
	inputs            *cwl.Values
	runtime           Runtime
	parentRuntime     Runtime
	fs                Filesystem
	inputFS, outputFS Filesystem
	bindings          []*Binding
	expressionLibs    []string
	env               map[string]string
	filesToCreate     []cwl.FileDir
	shell             bool
	resources         ResourcesLimites
	stdout            string
	stderr            string
	stdin             string
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
		if (!ok || any == nil) && in.GetInputParameter().Default != nil {
			(*values)[in.GetInputParameter().ID] = in.GetInputParameter().Default
		}
	}
}

func (process *Process) Command() ([]string, error) {
	// Copy "Tool.Inputs" bindings
	args := make([]*Binding, 0, len(process.bindings))
	// flat binding
	// process.FlatBinding()
	for _, b := range flatBinding(process.bindings, true) {
		if b.clb != nil {
			args = append(args, b)
		}
	}
	tool := process.root.Process.(*cwl.CommandLineTool)
	// Add "Tool.Arguments"
	for i, arg := range tool.Arguments {
		if arg.Binding == nil && arg.Exp != "" {
			expStr := string(arg.Exp)
			expResult, expErr := process.Eval(arg.Exp, nil)
			if expErr == nil {
				expStr = fmt.Sprint(expResult)
			}
			args = append(args, &Binding{
				arg.Binding, cwl.NewType(argType), expStr, sortKey{0}, nil, "",
			})
			continue
		} else if arg.Binding == nil {
			return nil, fmt.Errorf("empty argument")
		}
		if arg.Binding != nil && arg.Binding.ValueFrom == "" {
			return nil, fmt.Errorf("valueFrom is required but missing for argument %d", i)
		}
		args = append(args, &Binding{
			arg.Binding, cwl.NewType(argType), nil, sortKey{getPos(arg.Binding)}, nil, "",
		})
	}
	//
	// Evaluate "valueFrom" expression.
	for i, b := range args {
		if b.clb != nil && b.clb.ValueFrom != "" {
			val, err := process.eval(b.clb.ValueFrom, b.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to eval argument value: %s", err)
			}
			// 如果发生了文件赋值，可能会导致 migrateFile 失效，因此需要先备份;
			if b.Type.IsArray() {
				nested := &Binding{b.clb, *b.Type.MustArraySchema().GetItems(), val, b.sortKey, nil, b.name}
				args[i] = &Binding{b.clb, b.Type, val, b.sortKey, []*Binding{nested}, b.name}
			} else {
				args[i] = &Binding{b.clb, b.Type, val, b.sortKey, b.nested, b.name}
			}
			// b.Value = val
		}
	}

	sort.Stable(bySortKey(args))
	//debug(args)
	//
	//// Now collect the input bindings into command line arguments
	cmd := append([]string{}, tool.BaseCommands...)
	shellCommand := tool.RequiresShellCommand()
	for _, b := range args {
		cmd = append(cmd, bindArgs(b, shellCommand)...)
	}
	//
	if shellCommand {
		cmd = []string{"/bin/sh", "-c", strings.Join(cmd, " ")}
	}
	//debug("COMMAND", cmd)
	return cmd, nil
}

func flatBinding(nested []*Binding, checkClb bool) []*Binding {
	outs := make([]*Binding, 0, len(nested))
	for i, bi := range nested {
		var checked = bi.clb != nil

		if checkClb && checked {
			outs = append(outs, nested[i])
		} else if bi.nested != nil {
			outs = append(outs, flatBinding(nested[i].nested, checkClb)...)
		} else if !checkClb {
			outs = append(outs, nested[i])
		}
	}
	return outs
}

func (p *Process) loadRuntime() {
	p.vm.Set("runtime", p.runtime)
}

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
			binding = &cwl.CommandLineBinding{InputBinding: *in.InputBinding}
		}
		b, err := p.bindInput(in.ID, in.Type, binding, in.SecondaryFiles, val, k)
		if err != nil {
			return nil, p.errorf("binding input %q: %s", in.ID, err)
		}
		if b == nil {
			return nil, p.errorf("no binding found for input: %s", in.ID)
		}
		p.bindings = append(p.bindings, b...)
	}
	if err := p.initJVM(); err != nil {
		return nil, err
	}

	out, err := p.jsvm.Eval(tool.Expression, nil)
	//log.Printf("out %#v", out)
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

// 不同的 requirements 可能需要在不同的时机加载；可能需要拆分
func (process *Process) loadReqs() error {
	tool := process.root.Process.(*cwl.CommandLineTool)
	if req := tool.RequiresInlineJavascript(); req != nil {
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
		if err := process.bindEnvVar(req); err != nil {
			return err
		}
	}
	if req := tool.RequiresResource(); req != nil {
		process.loadRuntime()
	}
	if req := tool.RequiresSchemaDef(); req != nil {
		// 目前在 bindInput 中有用到时再进行的处理，可以考虑优化
	}
	return nil
}

func (process *Process) ResourcesLimites() (*ResourcesLimites, error) {
	limits := GetDefaultResourcesLimits()
	if req := process.tool.RequiresResource(); req != nil {
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

func (process *Process) RefreshVMInputs() error {
	return process.jsvm.setInputs(*process.inputs)
}
