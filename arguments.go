package cwl

import (
	"encoding/json"
)



func (p *Argument) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		p = nil
		return nil
	} else if data[0] == '{' {
		p.Binding = &CommandLineBinding{}
		return json.Unmarshal(data, p.Binding)
	}
	return json.Unmarshal(data, &p.Exp)
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
	return string(arg.Exp)
}

func (arg Argument) MustBinding() *CommandLineBinding {
	return arg.Binding
}

// Len for sorting.
func (args Arguments) Len() int {
	return len(args)
}

// Less for sorting.
func (args Arguments) Less(i, j int) bool {
	prev, next := args[i].Binding, args[j].Binding
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
