package cwl

import (
	"encoding/json"
)

//// Argument represents an element of "arguments" of CWL
//// @see http://www.commonwl.org/v1.0/CommandLineTool.html#CommandLineTool
//type Argument struct {
//	Value   string
//	Binding *Binding
//}

//// New constructs an "Argument" struct from any interface.
//func (_ Argument) New(i interface{}) Argument {
//	dest := Argument{}
//	switch x := i.(type) {
//	case string:
//		dest.Value = x
//	case map[string]interface{}:
//		dest.Binding = Binding{}.New(x)
//	}
//	return dest
//}

func (p *Argument)  UnmarshalJSON(data []byte) error{
	if len(data) == 0 {
		p = nil
		return nil
	} else if data[0] == '{' {
		p.binding = &CommandLineBinding{}
		return json.Unmarshal(data, p.binding)
	}
	return json.Unmarshal(data, &p.exp)
}

// Flatten ...
func (arg Argument) Flatten() []string {
	flattened := []string{}
	// TODO Do arg Flatten
	//if arg.Value != "" {
	//	flattened = append(flattened, arg.Value)
	//}
	//if arg.Binding != nil {
	//	if arg.Binding.Prefix != "" {
	//		flattened = append([]string{arg.Binding.Prefix}, flattened...)
	//	}
	//}
	return flattened
}

func (arg Argument) MustString() string {
	return string(arg.exp)
}

func (arg Argument) MustBinding() *CommandLineBinding {
	return arg.binding
}

// Len for sorting.
func (args Arguments) Len() int {
	return len(args)
}

// Less for sorting.
func (args Arguments) Less(i, j int) bool {
	prev, next := args[i].binding, args[j].binding
	switch [2]bool{prev == nil, next == nil} {
	case [2]bool{true, true}:
		return false
	case [2]bool{false, true}:
		return *prev.Position.Int < 0
	case [2]bool{true, false}:
		return *next.Position.Int > 0
	default:
		return *prev.Position.Int <= *next.Position.Int
	}
}

// Swap for sorting.
func (args Arguments) Swap(i, j int) {
	args[i], args[j] = args[j], args[i]
}
