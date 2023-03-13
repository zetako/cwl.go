package cwl_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/lijiang2014/cwl.go"
)

func TestCWL_tool_73(t *testing.T) {
	file, err := os.Open("cwl/v1.0/v1.0/record-output.cwl")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	pp, err := cwl.ParseCWLProcess(raw)
	if err != nil {
		t.Fatal(err)
	}
	p, ok := pp.(*cwl.CommandLineTool)
	if !ok {
		t.Fatalf("not CommandLineTool : %#v", p)
	}

	if p.ClassName() != "CommandLineTool" {
		t.Fatal("ClassName", p.ClassName())
	}
	pass := true
	pass = p.CWLVersion == "v1.0" &&
		len(p.BaseCommands) == 0 &&
		p.Arguments.Len() == 5 &&
		len(p.Inputs) == 1 && p.Inputs[0].GetInputParameter().ID == "irec" &&
		len(p.Outputs) == 1
	if !pass {
		t.Fatalf("%#v", p)
	}
	// ✅ Test Inputs
	input0 := p.Inputs[0].(*cwl.CommandInputParameter)
	b1 := input0.InputBinding
	if b1 != nil {
		t.Fatalf("input err inputbinding %#v", b1)
		t.Logf("%#v", b1)
	}
	if input0.Type.TypeName() != "record" {
		t.Fatalf("input err type %#v", input0.Type.TypeName())
	}
	inRec, got := input0.Type.MustRecord().(*cwl.CommandInputRecordSchema)
	if !got {
		t.Fatalf("%#v", inRec)
	}
	// ✅ Test inputs Type
	pass = inRec.Name == "irec" &&
		len(inRec.Fields) == 2
	if !pass {
		t.Fatalf("%s %#v", inRec.Name, inRec)
	}
	iRecIfoo := inRec.Fields[0]
	iRecIbar := inRec.Fields[1]
	pass = iRecIfoo.FieldName() == "ifoo" && iRecIfoo.FieldType().TypeName() == "File" &&
		iRecIbar.FieldName() == "ibar" && iRecIbar.FieldType().TypeName() == "File"
	if !pass {
		t.Fatalf("%v %#v", iRecIfoo, iRecIbar)
	}
	// iRecIfooFile := iRecIfoo.
	iRecIfooV, ok1 := iRecIfoo.(*cwl.CommandInputRecordField)
	iRecIbarV, ok2 := iRecIbar.(*cwl.CommandInputRecordField)
	if !ok1 || !ok2 {
		t.Fatalf("%#v %#v", iRecIfoo, iRecIbar)
	}
	pass = iRecIfooV.InputBinding != nil && iRecIfooV.InputBinding.Position.MustInt() == 2 &&
		iRecIbarV.InputBinding != nil && iRecIbarV.InputBinding.Position.MustInt() == 6
	if !pass {
		t.Fatalf("%#v %#v", iRecIfooV, iRecIbarV)
	}

	// ✅ Test Arguments
	args := p.Arguments
	pass = args.Len() == 5
	if !pass {
		t.Fatalf("arguments err  %#v", args)
	}
	arg0 := args[0].MustBinding()
	arg1 := args[1].MustBinding()
	arg2 := args[2].MustBinding()
	arg3 := args[3].MustBinding()
	arg4 := args[4].MustBinding()
	pass = arg0 != nil && *arg0.Position.Int == 1 && arg0.ValueFrom == "cat" &&
		arg1 != nil && *arg1.Position.Int == 3 && arg1.ValueFrom == "> foo" && !arg1.ShellQuote &&
		arg2 != nil && *arg2.Position.Int == 4 && arg2.ValueFrom == "&&" && !arg2.ShellQuote &&
		arg3 != nil && *arg3.Position.Int == 5 && arg3.ValueFrom == "cat" &&
		arg4 != nil && *arg4.Position.Int == 7 && arg4.ValueFrom == "> bar" && !arg4.ShellQuote

	if !pass {
		t.Fatalf("arguments err binding %#v %#v %#v %#v %#v", arg0, arg1, arg2, arg3, arg4)
	}
	// ✅ Test Requirements
	reqs := p.Requirements
	_, ok0 := reqs[0].(*cwl.ShellCommandRequirement)
	pass = len(reqs) == 1 && ok0
	if !pass {
		t.Fatalf("Requirements check failed %#v", p.Requirements)
	}

	o0 := p.Outputs[0].(*cwl.CommandOutputParameter)
	pass = o0 != nil &&
		o0.ID == "orec" && o0.OutputBinding == nil
	if !pass {
		t.Fatalf("out err  %#v", o0)
	}
	//  Test Output Type
	ot0 := o0.Type
	//t.Logf("%#v",ot0.TypeName())
	pass = ot0.TypeName() == "record"
	if !pass {
		t.Fatalf("out type err  %#v", ot0)
	}
	// inRec, got := input0.Type.MustRecord().(*cwl.CommandInputRecordSchema)
	orec, got := ot0.MustRecord().(*cwl.CommandOutputRecordSchema)
	if !got || len(orec.Fields) != 2 {
		t.Fatalf("out type err  %#v", ot0)
	}
	ofoo := orec.Fields[0]
	obar := orec.Fields[1]
	pass = ofoo.FieldName() == "ofoo" && ofoo.FieldType().TypeName() == "File" &&
		obar.FieldName() == "obar" && obar.FieldType().TypeName() == "File"
	if !pass {
		t.Fatalf("out0 type err  %#v", ot0)
	}
	ofooV := ofoo.(*cwl.CommandOutputRecordField)
	obarV := obar.(*cwl.CommandOutputRecordField)
	pass = ofooV.OutputBinding.Glob[0] == "foo" && obarV.OutputBinding.Glob[0] == "bar"
	if !pass {
		t.Fatalf("out err  %#v %#v", ofoo, obar)
	}
}
