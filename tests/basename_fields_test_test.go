package cwlgotest

import (
	"sort"
	"testing"

	cwl "github.com/lijiang2014/cwl.go"
	. "github.com/otiai10/mint"
)

func TestDecode_basename_fields_test(t *testing.T) {
	f := load("basename-fields-test.cwl")
	doc := cwl.NewCWL()
	err := doc.Decode(f)
	root := doc.Process.(*cwl.Workflow)
	Expect(t, err).ToBe(nil)
	Expect(t, root.CWLVersion).ToBe("v1.0")
	Expect(t, root.Class).ToBe("Workflow")
	Expect(t, root.Requirements[0].ClassName()).ToBe("StepInputExpressionRequirement")

	sort.Sort(root.Inputs)
	Expect(t, root.Inputs[0].(*cwl.WorkflowInputParameter).ID).ToBe("tool")
	Expect(t, root.Inputs[0].(*cwl.WorkflowInputParameter).Type.TypeName()).ToBe("File")
	// TODO check specification for this test ID and Type
	Expect(t, root.Outputs[0].(*cwl.WorkflowOutputParameter).ID).ToBe("extFile")
	Expect(t, root.Outputs[0].(*cwl.WorkflowOutputParameter).Type.TypeName()).ToBe("File")
	Expect(t, root.Outputs[0].(*cwl.WorkflowOutputParameter).OutputSource[0]).ToBe("ext/out")
	Expect(t, root.Outputs[1].(*cwl.WorkflowOutputParameter).ID).ToBe("rootFile")
	Expect(t, root.Outputs[1].(*cwl.WorkflowOutputParameter).Type.TypeName()).ToBe("File")
	Expect(t, root.Outputs[1].(*cwl.WorkflowOutputParameter).OutputSource[0]).ToBe("root/out")
	count := 0
	for _, st := range root.Steps {
		switch st.ID {
		case "root":
			Expect(t, st.Run.ID).ToBe("echo-file-tool.cwl")
			//t.Logf("%#v",st.In[0])
			Expect(t, string(st.In[0].ValueFrom)).ToBe("$(inputs.tool.nameroot)")
			Expect(t, st.Out[0].ID).ToBe("out")
			// TODO tool: tool
			count = count + 1
		case "ext":
			Expect(t, st.Run.ID).ToBe("echo-file-tool.cwl")
			Expect(t, string(st.In[0].ValueFrom)).ToBe("$(inputs.tool.nameext)")
			Expect(t, st.Out[0].ID).ToBe("out")
			// TODO tool: tool
			count = count + 1
		}
	}
	Expect(t, count).ToBe(2)
}
