package runner

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lijiang2014/cwl.go"
	"github.com/robertkrimen/otto"
)

type jsvm struct {
	inputsData map[string]interface{}
	runtime    map[string]interface{}
	// expressionLibs
	vm *otto.Otto
}

func (process *Process) initJVM() error {
	var vm *jsvm
	if process.jsvm == nil {
		vm = &jsvm{}
		vm.vm = otto.New()
		process.jsvm = vm
	} else {
		vm = process.jsvm
	}
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
	vm.vm.Set("inputs", inputsData)
	vm.vm.Set("runtime", vm.runtime)
	return nil
}

func (vm *jsvm) Eval(e cwl.Expression, data interface{}) (interface{}, error) {
	vm.vm.Set("self", data)
	return vm.EvalParts(parseExp(e))
}

type ExpPart struct {
	Raw  string
	Expr string
	//Start, End int
	// true if the expression is a javascript function body (e.g. ${return "foo"})
	IsFuncBody bool
}

// See https://www.commonwl.org/v1.2/CommandLineTool.html#Expressions_(Optional)
// var rx = regexp.MustCompile(`\$\((.*)\)`)
// Parameter references
var rx = regexp.MustCompile(`\$\([\pL\pN]+(\.[\pL\pN]+|\['[^'| ]+'\]|\["[^"| ]+"\]|\[\d+\])*\)`)

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
				Raw:  e,
				Expr: strings.TrimSpace(ev[2 : len(ev)-1]),
				//Start:      0,
				//End:        len(e),
				IsFuncBody: true,
			},
		}
	}

	var parts []*ExpPart
	matches := CwlExprSacner(e)
	for _, match := range matches {
		parts = append(parts, &ExpPart{
			Raw:        match[0],
			Expr:       match[1] + match[2],
			IsFuncBody: match[2] != "",
		})
	}
	// parse parameter reference
	//last := 0
	//matches := rx.FindAllStringSubmatchIndex(e, -1)
	//for _, match := range matches {
	//	start := match[0]
	//	end := match[1]
	//	gstart := match[2]
	//	gend := match[3]
	//
	//	if start > last {
	//		parts = append(parts, &ExpPart{
	//			Raw:   e[last:start],
	//			Start: last,
	//			End:   start,
	//		})
	//	}
	//
	//	parts = append(parts, &ExpPart{
	//		Raw:   string(e[start:end]),
	//		Expr:  string(e[gstart:gend]),
	//		Start: start,
	//		End:   end,
	//	})
	//	last = end
	//}
	//
	//if last < len(e)-1 {
	//	parts = append(parts, &ExpPart{
	//		Raw:   string(e[last:]),
	//		Start: last,
	//		End:   len(e),
	//	})
	//}

	return parts
}

func (vm *jsvm) EvalParts(parts []*ExpPart) (interface{}, error) {

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

		val, err := vm.vm.Run(code)
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

			val, err := vm.vm.Run(part.Expr)
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

func CwlExprSacner(in string) [][3]string {
	var pass, l = 0, len(in)
	var str, exp, funCode string
	var inExp, inFun bool
	var deep int
	ret := make([][3]string, 0)
	for i, c := range in {
		if pass > 0 {
			pass--
			continue
		}
		// scan 3 byte
		if !inExp && !inFun {
			if l-i < 2 {
				// do nothing
			} else if c == '\\' {
				if in[i:i+2] == "\\$(" {
					pass = 2
					str += "$("
					continue
				}
				if in[i:i+2] == "\\${" {
					pass = 2
					str += "${"
					continue
				}
				if in[i:i+2] == `\\` {
					pass = 1
					str += "\\"
					continue
				}
			} else if c == '$' {
				if in[i+1] == '(' {
					pass = 1
					inExp = true
					if str != "" {
						ret = append(ret, [3]string{str, "", ""})
						str = ""
					}
					continue
				} else if in[i+1] == '{' {
					pass = 1
					inFun = true
					if str != "" {
						ret = append(ret, [3]string{str, "", ""})
						str = ""
					}
					continue
				}
			}
		}
		if inExp {
			if c == '(' {
				deep++
			} else if c == ')' {
				if deep == 0 {
					// close Exp
					ret = append(ret, [3]string{str, exp, funCode})
					str, exp, funCode = "", "", ""
					inExp = false
					continue
				}
				deep--
			}
			exp = exp + in[i:i+1]
		} else if inFun {
			if c == '{' {
				deep++
			} else if c == '}' {
				if deep == 0 {
					// close Exp
					ret = append(ret, [3]string{str, exp, funCode})
					str, exp, funCode = "", "", ""
					inFun = false
					continue
				}
				deep--
			}
			funCode = funCode + in[i:i+1]
		} else {
			str += in[i : i+1]
		}
	}
	if str != "" {
		ret = append(ret, [3]string{str, exp, funCode})
	}
	return ret
}
