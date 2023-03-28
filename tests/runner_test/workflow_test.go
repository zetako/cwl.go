package runnertest

import (
	irunner "github.com/lijiang2014/cwl.go/runner"
	"testing"

	. "github.com/otiai10/mint"
)

// workflow
// ❌ any-type-compat.cwl v1.0/any-type-job.json
func TestCWLR2Workflow_run20(t *testing.T) {
	testByID(t, 20)
}

func TestCWLWorkflow_workon(t *testing.T) {
	e, err := newEngine("v1.0/count-lines1-wf.cwl", "v1.0/wc-job.json") // 1
	Expect(t, err).ToBe(nil)
	t.Log("Break Here")
	e.SetDefaultExecutor(&irunner.LocalExecutor{})
	_, err = e.MainProcess()
	Expect(t, err).ToBe(nil)
	t.Log("Break Here")
	outs, err := e.Run()
	Expect(t, err).ToBe(nil)
	t.Log(outs)
	t.Log("Break Here")
	//t.Logf("%#v", p.Root().Process)
	//tool := p.Root().Process.(*cwl.Workflow)
	//_ = tool
	//for i, ini := range tool.Inputs {
	//	in := ini.(*cwl.WorkflowInputParameter)
	//	t.Logf("%d %s %s", i, in.ID, in.Type.TypeName())
	//	//if in.Type.IsArray() {
	//	//  t.Logf("%#v", in.Type.MustArraySchema())
	//	//}
	//}
	//t.Logf("%#v", tool.Outputs)
	//for i, vali := range tool.Outputs {
	//	t.Logf("%d %#v", i, vali)
	//	val := vali.(*cwl.WorkflowOutputParameter)
	//	if val.Type.IsArray() {
	//		t.Logf("%#v", val.Type.MustArraySchema())
	//	}
	//	if val.OutputSource != nil {
	//		t.Logf("%#v", val.OutputSource)
	//	}
	//}
	//t.Logf("%#v", tool.Requirements)
	//for i, vali := range tool.Requirements {
	//	t.Logf("%d %#v", i, vali)
	//}
	// err = os.RemoveAll("/tmp/testcwl")
}
