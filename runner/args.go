package runner

import (
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"strings"
)

// bySortKey defines the rules for sorting bindings;
// http://www.commonwl.org/v1.0/CommandLineTool.html#Input_binding
type bySortKey []*Binding

func (s bySortKey) Len() int      { return len(s) }
func (s bySortKey) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s bySortKey) Less(i, j int) bool {
	z := compareKey(s[i].sortKey, s[j].sortKey)
	// cwl spec
	//  If and only if two bindings have the same sort key,
	// the tie must be broken using the ordering of the field or parameter name
	// immediately containing the leaf binding.
	if z == 0 {
		return s[i].name < s[j].name
	}
	return z == -1
}

// compare two sort keys.
//
// The result will be 0 if i==j, -1 if i < j, and +1 if i > j.
func compareKey(i, j sortKey) int {
	for x := 0; x < len(i) || x < len(j); x++ {
		if x >= len(i) {
			// i key is shorter than j
			return -1
		}
		if x >= len(j) {
			// j key is shorter than i
			return 1
		}
		z := compare(i[x], j[x])
		if z != 0 {
			return z
		}
	}
	return 0
}

// compare two sort key items, because sort keys may have mixed ints and strings.
// cwl spec: "ints sort before strings", i.e all ints are less than all strings.
//
// The result will be 0 if i==j, -1 if i < j, and +1 if i > j.
func compare(iv, jv interface{}) int {
	istr, istrok := iv.(string)
	jstr, jstrok := jv.(string)
	iint, iintok := iv.(int)
	jint, jintok := jv.(int)

	switch {
	case istrok && jintok:
		// i is a string, j is an int
		// cwl spec: "ints sort before strings"
		return 1
	case iintok && jstrok:
		// i is an int, j is a string
		// cwl spec: "ints sort before strings"
		return -1

	// both are strings
	case istrok && jstrok && istr == jstr:
		return 0
	case istrok && jstrok && istr < jstr:
		return -1
	case istrok && jstrok && istr > jstr:
		return 1

	// both are ints
	case iintok && jintok && iint == jint:
		return 0
	case iintok && jintok && iint < jint:
		return -1
	case iintok && jintok && iint > jint:
		return 1
	}
	return 0
}

func getPos(in *cwl.CommandLineBinding) int {
	if in == nil {
		return 0
	}
	if in.Position == nil {
		return 0
	}
	ret, _ := in.Position.Value()
	return int(ret)
}

// args converts a binding into a list of formatted command line arguments.
func bindArgs(b *Binding) []string {
	switch b.Type.TypeName() {

	case "array":
		// cwl conformance test:
		// Test [67/68] Test that empty array input does not add anything to command line
		// TODO unhandled panic here
		arr := b.Value.([]cwl.Value)
		if len(arr) == 0 {
			return nil
		}

		// cwl spec:
		// "If itemSeparator is specified, add prefix and the join the array
		// into a single string with itemSeparator separating the items..."
		if b.clb != nil && b.clb.ItemSeparator != "" {

			var nested []cwl.Value
			for _, nb := range b.nested {
				nested = append(nested, nb.Value)
			}
			return formatArgs(b.clb, nested...)

			// cwl spec:
			// "...otherwise first add prefix, then recursively process individual elements."
		} else {
			args := formatArgs(b.clb)

			for _, nb := range b.nested {
				args = append(args, bindArgs(nb)...)
			}
			return args
		}

	case "record":
		// TODO

	case "any", "string", "int", "long", "float", "double", "File", "Directory", argType:
		//  case cwl.Any, cwl.String, cwl.Int, cwl.Long, cwl.Float, cwl.Double, cwl.FileType,
		//cwl.DirectoryType, argType:
		return formatArgs(b.clb, b.Value)

	case "bool":
		// cwl spec:
		// "boolean: If true, add prefix to the command line. If false, add nothing."
		bv := b.Value.(bool)
		if bv && b.clb != nil && b.clb.Prefix != "" {
			return formatArgs(b.clb)
		}
	}
	return nil
}

// formatArgs applies some command line binding rules to a CLI argument,
// such as prefix, separate, etc.
// http://www.commonwl.org/v1.0/CommandLineTool.html#CommandLineBinding
func formatArgs(clb *cwl.CommandLineBinding, args ...cwl.Value) []string {
	var (
		prefix, join string
		sep          = true
		strargs      []string
	)
	if clb != nil {
		prefix = clb.Prefix
		join = clb.ItemSeparator
		sep = clb.Separate
	}

	for _, arg := range args {
		strargs = append(strargs, valueToStrings(arg)...)
	}

	if join != "" && strargs != nil {
		strargs = []string{strings.Join(strargs, join)}
	}

	if prefix != "" {
		if sep {
			strargs = append([]string{prefix}, strargs...)
		} else if strargs != nil {
			strargs[0] = prefix + strargs[0]
		} else {
			strargs = []string{prefix}
		}
	}
	return strargs
}

func valueToStrings(v cwl.Value) []string {
	switch z := v.(type) {
	case []interface{}:
		var out []string
		for _, v := range z {
			out = append(out, valueToStrings(v)...)
		}
		return out
	case int, int32, int64, float32, float64, bool, string:
		return []string{fmt.Sprintf("%v", z)}
	case cwl.File:
		return []string{z.Path}
	}
	return nil
}
