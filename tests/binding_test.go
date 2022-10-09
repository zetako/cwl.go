package cwlgotest

import (
	"sort"
	"testing"
	
	"github.com/lijiang2014/cwl.go"
	. "github.com/otiai10/mint"
)

func TestDecode_binding_test(t *testing.T) {
	f := load("binding-test.cwl")
	doc := cwl.NewCWL()
	err := doc.Decode(f)
	root := doc.Process.(*cwl.CommandLineTool)
	Expect(t, err).ToBe(nil)

	Expect(t, root.CWLVersion).ToBe("v1.0")
	Expect(t, root.Class).ToBe("CommandLineTool")

	Expect(t, root.Hints[0].ClassName()).ToBe("DockerRequirement")
	Expect(t, root.Hints[0].(*cwl.DockerRequirement).DockerPull).ToBe("python:2-slim")
	//
	sort.Sort(root.Inputs)
	input0 := root.Inputs[0].(*cwl.CommandInputParameter)
	Expect(t, input0.ID).ToBe("#args.py")
	Expect(t, input0.Type.TypeName()).ToBe("File")
	Expect(t, input0.Default.(map[string]interface{})["class"]).ToBe("File")
	Expect(t, int(*input0.InputBinding.Position.Int)).ToBe(-1)
	
	input1 := root.Inputs[1].(*cwl.CommandInputParameter)
	Expect(t, input1.ID).ToBe("reads")
	Expect(t, input1.Type.TypeName()).ToBe("array")
	Expect(t, input1.Type.MustArraySchema().Items.TypeName()).ToBe("File")
	Expect(t, input1.Type.Binding.Prefix).ToBe("-YYY")
	Expect(t, input1.InputBinding.Position.MustInt()).ToBe(3)
	Expect(t, input1.InputBinding.Prefix).ToBe("-XXX")
	
	input2 := root.Inputs[2].(*cwl.CommandInputParameter)
	Expect(t, input2.ID).ToBe("reference")
	Expect(t, input2.Type.TypeName()).ToBe("File")
	Expect(t, int(*input2.InputBinding.Position.Int)).ToBe(2)
	
	
	output0 := root.Outputs[0].(*cwl.CommandOutputParameter)
	t.Logf("%#v",output0)
	Expect(t, output0.ID).ToBe("args")
	Expect(t, output0.Type.ShortString()).ToBe("string[]")
}
