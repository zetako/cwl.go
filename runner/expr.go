package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/robertkrimen/otto"
	"regexp"
	"strings"
)

type jsvm struct {
	inputsData map[string]interface{}
	runtime    map[string]interface{}
	// expressionLibs
	vm *otto.Otto
}

func (process *Process) initJVM() error {
	vm := &jsvm{}
	inputsData := map[string]interface{}{}
	for _, b := range process.bindings {
		v, err := toJSONMap(b.Value)
		if err != nil {
			return fmt.Errorf(`mashaling "%s" for JS eval`, b.name)
		}
		if v == nil {
			v = otto.NullValue()
		}
		inputsData[b.name] = v
	}
	vm.inputsData = inputsData
	r := process.runtime
	vm.runtime = map[string]interface{}{
		"outdir":     r.Outdir,
		"tmpdir":     r.Tmpdir,
		"cores":      r.Cores,
		"ram":        r.RAM,
		"outdirSize": r.OutdirSize,
		"tmpdirSize": r.TmpdirSize,
	}
	process.jsvm = vm
	vm.vm = otto.New()
	vm.vm.Set("inputs", inputsData)
	vm.vm.Set("runtime", vm.runtime)
	return nil
}

func (vm *jsvm) Eval(e cwl.Expression, data interface{}) (interface{}, error) {
	vm.vm.Set("self", data)
	return vm.EvalParts(parseExp(e))
}

type ExpPart struct {
	Raw        string
	Expr       string
	Start, End int
	// true if the expression is a javascript function body (e.g. ${return "foo"})
	IsFuncBody bool
}

var rx = regexp.MustCompile(`\$\((.*)\)`)

func parseExp(expr cwl.Expression) []*ExpPart {
	e := string(expr)
	ev := strings.TrimSpace(e)
	if len(ev) == 0 {
		return nil
	}

	// javascript function expression
	if strings.HasPrefix(ev, "${") && strings.HasSuffix(ev, "}") {
		return []*ExpPart{
			{
				Raw:        e,
				Expr:       strings.TrimSpace(ev[2 : len(ev)-1]),
				Start:      0,
				End:        len(e),
				IsFuncBody: true,
			},
		}
	}

	var parts []*ExpPart

	// parse parameter reference
	last := 0
	matches := rx.FindAllStringSubmatchIndex(e, -1)
	for _, match := range matches {
		start := match[0]
		end := match[1]
		gstart := match[2]
		gend := match[3]

		if start > last {
			parts = append(parts, &ExpPart{
				Raw:   e[last:start],
				Start: last,
				End:   start,
			})
		}

		parts = append(parts, &ExpPart{
			Raw:   string(e[start:end]),
			Expr:  string(e[gstart:gend]),
			Start: start,
			End:   end,
		})
		last = end
	}

	if last < len(e)-1 {
		parts = append(parts, &ExpPart{
			Raw:   string(e[last:]),
			Start: last,
			End:   len(e),
		})
	}

	return parts
}

func (j *jsvm) EvalParts(parts []*ExpPart) (interface{}, error) {

	var vm = j.vm

	if len(parts) == 0 {
		return nil, nil
	}
	if len(parts) == 1 {
		part := parts[0]

		// No expression, just a normal string.
		if part.Expr == "" {
			return part.Raw, nil
		}

		// Expression or JS function body.
		// Can return any type.
		//code := strings.Join(libs, "\n")
		code := ""
		if part.IsFuncBody {
			code = "(function(){" + part.Expr + "})()"
		} else {
			code = "(function(){ return " + part.Expr + "; })()"
		}

		val, err := vm.Run(code)
		if err != nil {
			return nil, fmt.Errorf("failed to run JS expression: %s", err)
		}

		// otto docs:
		// "Export returns an error, but it will always be nil.
		//  It is present for backwards compatibility."
		ival, _ := val.Export()
		return ival, nil
	}

	// There are multiple parts for expressions of the form "foo $(bar) baz"
	// which is to be treated as string interpolation.

	res := ""
	for _, part := range parts {
		if part.Expr != "" {

			val, err := vm.Run(part.Expr)
			if err != nil {
				return nil, fmt.Errorf("failed to run JS expression: %s", err)
			}

			sval, err := val.ToString()
			if err != nil {
				return nil, fmt.Errorf("failed to convert JS result to a string: %s", err)
			}

			res += sval
		} else {
			res += part.Raw
		}
	}
	return res, nil
}
