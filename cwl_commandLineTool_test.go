package cwl_test

import (
	"github.com/lijiang2014/cwl.go"
	"io/ioutil"
	"os"
	"testing"
)

func TestCWL_tool_1(t *testing.T) {
	file, err := os.Open("cwl/v1.0/v1.0/bwa-mem-tool.cwl")
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
		len(p.BaseCommands) == 1 && p.BaseCommands[0] == "python" &&
		//p.Arguments.Len() == 3 &&
		len(p.Inputs) == 5 && p.Inputs[0].GetInputParameter().ID == "reference" &&
		len(p.Outputs) == 2 &&
		p.Outputs[0].GetOutputParameter().ID == "sam" &&
		p.Outputs[1].GetOutputParameter().ID == "args" &&
		p.Stdout == "output.sam"
	if !pass {
		t.Fatalf("%#v", p)
	}
	// ✅ Test Inputs
	input0 := p.Inputs[0].(*cwl.CommandInputParameter)
	b1 := input0.InputBinding
	pass = b1 != nil &&
		b1.Position != nil &&
		b1.Position.Int != nil &&
		*b1.Position.Int == 2 &&
		b1.Separate && b1.ShellQuote
	if !pass {
		t.Fatalf("input err inputbinding %#v", b1)
		t.Logf("%#v", b1)
	}
	// ✅ Test inputs Type
	pass = input0.Type.TypeName() == "File" &&
		input0.Type.IsNullable() == false
	if !pass {
		t.Fatalf("input0 type err%#v", input0.Type)
	}
	typein1 := p.Inputs[1].(*cwl.CommandInputParameter).Type
	pass = typein1.TypeName() == "array" &&
		typein1.IsArray() == true && typein1.MustArraySchema().GetItems().TypeName() == "File"
	if !pass {
		t.Fatalf("input1 type err%#v %#v %s %s", typein1, typein1.MustArraySchema(), typein1.TypeName(), typein1.MustArraySchema().GetItems().TypeName())
	}
	typein2 := p.Inputs[2].(*cwl.CommandInputParameter).Type
	pass = typein2.TypeName() == "int" &&
		cwl.IsPrimitiveSaladType(typein2.TypeName())
	if !pass {
		t.Fatalf("input2 type err%#v", typein2)
	}
	typein3 := p.Inputs[3].(*cwl.CommandInputParameter).Type
	pass = typein3.TypeName() == "array" &&
		typein3.IsArray() == true && typein3.MustArraySchema().GetItems().TypeName() == "int"
	if !pass {
		t.Fatalf("input3 type err%#v %s", typein3, typein3.TypeName())
	}
	typein4 := p.Inputs[4].(*cwl.CommandInputParameter).Type
	pass = typein4.TypeName() == "File"
	if !pass {
		t.Fatalf("input3 type err%#v %s", typein4, typein4.TypeName())
	}
	//  Test input default , more
	input4 := p.Inputs[4].(*cwl.CommandInputParameter)
	t.Logf("default %#v",input4.Default)
	if input4.Default == nil {
		t.Fatalf("input4 default fail %#v", input4)
	}
	_ , ok = input4.Default.(cwl.File)
	if !ok {
		t.Fatalf("input4 default is not File Type %#v", input4)
	}
	// ✅ Test Arguments
	args := p.Arguments
	pass = args.Len() == 3 &&
		args[0].MustString() == "bwa" &&
		args[1].MustString() == "mem"
	if !pass {
		t.Fatalf("arguments err  %#v", args)
		t.Logf("%#v", b1)
	}
	arg3 := args[2].MustBinding()
	pass = arg3 != nil &&
		*arg3.Position.Int == 1 &&
		arg3.Prefix == "-t" &&
		arg3.ValueFrom == "$(runtime.cores)"
	if !pass {
		t.Fatalf("arguments err binding %#v", arg3)
		t.Logf("%#v", b1)
	}
	// ✅ Test Requirements
	reqs := p.Requirements
	reqs0, ok0 := reqs[0].(*cwl.ResourceRequirement)
	reqs1, ok1 := reqs[1].(*cwl.DockerRequirement)
	pass = len(reqs) == 2 && ok0 && ok1 &&
		reqs[0].ClassName() == "ResourceRequirement" &&
		reqs[1].ClassName() == "DockerRequirement"
	if !pass {
		t.Fatalf("Requirements check failed %#v", p.Requirements)
	}
	pass = *reqs0.CoresMin.Long == 2
	if !pass {
		t.Fatalf("Requirements 0 check failed %#v", reqs[0])
	}
	pass = reqs1.DockerPull == "python:2-slim"
	if !pass {
		t.Fatalf("Requirements 0 check failed %#v", reqs[0])
	}
	o0 := p.Outputs[0].(*cwl.CommandOutputParameter)
	pass = o0 != nil &&
		o0.ID == "sam" && o0.OutputBinding != nil
	if !pass {
		t.Fatalf("out err  %#v", o0)
	}
	ob1 := o0.OutputBinding
	pass = len(ob1.Glob) == 1 && ob1.Glob[0] == "output.sam"
	if !pass {
		t.Fatalf("out binding err  %#v", o0)
		t.Logf("%#v", ob1)
	}
	// TODO Test Output Type
	ot0 := o0.Type
	//t.Logf("%#v",ot0.TypeName())
	pass = ot0.IsNullable() && len(ot0.MustMulti()) == 2
	if !pass {
		t.Fatalf("out type err  %#v", ot0)
	}
	ot0t0 := ot0.MustMulti()[0]
	ot0t1 := ot0.MustMulti()[1]
	pass = ot0t0.TypeName() == "File" || ot0t1.TypeName() == "File"
	if !pass {
		t.Fatalf("out0 type err  %#v", ot0)
	}
	ot2 := p.Outputs[1].(*cwl.CommandOutputParameter).Type
	pass = ot2.TypeName() == "array" && ot2.MustArraySchema().GetItems().TypeName() == "string"
	if !pass {
		t.Fatalf("out1 type err  %#v", ot2)
	}
}

