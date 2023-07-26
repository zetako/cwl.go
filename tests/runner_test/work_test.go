package runnertest

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/runner"
	. "github.com/otiai10/mint"
)

func TestCWLR2_workon(t *testing.T) {
	e, err := newEngine("v1.0/bwa-mem-tool.cwl", "v1.0/bwa-mem-job.json") // 1
	Expect(t, err).ToBe(nil)
	p, err := e.MainProcess()
	Expect(t, err).ToBe(nil)
	//t.Logf("%#v", p.Root().Process)
	tool := p.Root().Process.(*cwl.CommandLineTool)
	_ = tool
	for i, ini := range tool.Inputs {
		in := ini.(*cwl.CommandInputParameter)
		t.Logf("%d %s %s", i, in.ID, in.Type.TypeName())
		//if in.Type.IsArray() {
		//  t.Logf("%#v", in.Type.MustArraySchema())
		//}
	}
	//t.Logf("%#v",tool.Outputs)
	//for i, vali := range tool.Outputs {
	// t.Logf("%d %#v", i, vali)
	// val := vali.(*cwl.CommandOutputParameter)
	// if val.Type.IsArray() {
	//   t.Logf("%#v", val.Type.MustArraySchema())
	// }
	//  if val.OutputBinding != nil {
	//    t.Logf("%#v", val.OutputBinding)
	//  }
	//}
	//t.Logf("%#v",tool.Requirements)
	//for i, vali := range tool.Requirements {
	//t.Logf("%d %#v", i, vali)
	//schema , ok:= vali.(*cwl.SchemaDefRequirement)
	//if ok {
	//  for j, typej := range schema.Types {
	//    input1 := typej.(*cwl.CommandInputType)
	//    t.Logf(" %d %#v",j, input1.TypeName())
	//    if input1.TypeName() == "record" {
	//      record := input1.MustRecord().(*cwl.CommandInputRecordSchema)
	//      t.Logf("%d %s %#v",j,  record.Name, record.InputBinding)
	//      for m , fieldm := range record.Fields {
	//        fi := fieldm.(*cwl.CommandInputRecordField)
	//        t.Logf(" %d %s  %s %s",m, fi.Name , fi.Type.TypeName(), fi.Type.String())
	//      }
	//    }
	//  }
	//}
	//}
	err = os.RemoveAll("/tmp/testcwl")

	limits, err := p.ResourcesLimites()
	Expect(t, err).ToBe(nil)
	runtime := L2R(*limits)
	_ = p.SetRuntime(runtime)
	err = e.ResolveProcess(p)
	Expect(t, err).ToBe(nil)
	cmds, err := p.Command()
	Expect(t, err).ToBe(nil)
	log.Println(cmds)
	log.Println("std:", tool.Stdin, tool.Stdout, tool.Stderr)
	// tmap mapall stage1 map1
	// --min-seq-length 20 map2 --min-seq-length 20 stage2 map1 --max-seq-length 20 --min-seq-length 10 --seed-length 16 map2 --max-seed-hits -1 --max-seq-length 20 --min-seq-length 10
	for i, args := range cmds {
		log.Println(i, args)
	}
	envs := p.Env()
	for k, v := range envs {
		log.Printf("SET %s=%s", k, v)
	}
	ex := runner.LocalExecutor{}
	// Expect(t, err).ToBe(nil)
	pid, ret, err := ex.Run(p)
	Expect(t, err).ToBe(nil)
	t.Log(pid)
	retCode, _ := <-ret
	Expect(t, retCode).ToBe(0)
	outputs, err := e.Outputs()
	Expect(t, err).ToBe(nil)
	t.Logf("%#v", outputs)
	for key, outi := range outputs {
		t.Logf("%s: %#v", key, outi)
	}
	raw, _ := json.Marshal(outputs)
	t.Logf("%s", string(raw))

}

func TestCWLR2_workon_expression(t *testing.T) {
	e, err := newEngine("v1.0/parseInt-tool.cwl", "v1.0/parseInt-job.json")
	Expect(t, err).ToBe(nil)
	p, err := e.MainProcess()
	Expect(t, err).ToBe(nil)
	//t.Logf("%#v", p.Root().Process)
	tool := p.Root().Process.(*cwl.ExpressionTool)
	_ = tool
	for i, ini := range tool.Inputs {
		in := ini.(*cwl.WorkflowInputParameter)
		t.Logf("%d %s %s", i, in.ID, in.Type.TypeName())
		//if in.Type.IsArray() {
		//  t.Logf("%#v", in.Type.MustArraySchema())
		//}
	}
	t.Logf("%#v", tool.Outputs)
	for i, vali := range tool.Outputs {
		t.Logf("%d %#v", i, vali)
		val := vali.(*cwl.ExpressionToolOutputParameter)
		if val.Type.IsArray() {
			t.Logf("%#v", val.Type.MustArraySchema())
		}
	}
	t.Logf("%#v", tool.Requirements)
	for i, vali := range tool.Requirements {
		t.Logf("%d %#v", i, vali)
	}

	outputs, err := p.RunExpression()
	Expect(t, err).ToBe(nil)
	t.Log(outputs)

	//pid, ret, err := ex.Run(p)
	//t.Log(pid)
	//retCode, _ := <-ret
	//Expect(t, retCode).ToBe(0)
	//outputs, err := e.Outputs()
	//Expect(t, err).ToBe(nil)
	//for key, outi := range outputs {
	//  t.Logf("%s: %#v", key, outi)
	//}
}
