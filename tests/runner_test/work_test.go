package runnertest

import (
  "github.com/lijiang2014/cwl.go"
  . "github.com/otiai10/mint"
  "log"
  "testing"
)

func TestCWLR2_workon(t *testing.T) {
  e, err := newEngine("v1.0/template-tool.cwl", "v1.0/cat-job.json")
  Expect(t, err).ToBe(nil)
  p, err := e.MainProcess()
  Expect(t, err).ToBe(nil)
  //t.Logf("%#v", p.Root().Process)
  tool := p.Root().Process.(*cwl.CommandLineTool)
  _ =tool
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
  limits, err := p.ResourcesLimites()
  Expect(t, err).ToBe(nil)
  runtime := L2R(*limits)
  p.SetRuntime(runtime)
  err = e.ResolveProcess(p)
  Expect(t, err).ToBe(nil)
  cmds, err := p.Command()
  Expect(t, err).ToBe(nil)
  log.Println(cmds)
  // tmap mapall stage1 map1
  // --min-seq-length 20 map2 --min-seq-length 20 stage2 map1 --max-seq-length 20 --min-seq-length 10 --seed-length 16 map2 --max-seed-hits -1 --max-seq-length 20 --min-seq-length 10
  for i, args := range cmds {
    log.Println(i, args)
  }
  //Expect(t, len(cmds)).ToBe(13)
  ////Expect(t, cmds[3]).ToBe(["bwa","mem"])
  //checks := []string{"bwa", "mem", "-t", "2", "-I", "1,2,3,4", "-m", "3", "chr20.fa",
  //  "example_human_Illumina.pe_1.fastq",
  //  "example_human_Illumina.pe_2.fastq"}
  //ExpectArray(t, cmds[2:10], checks[:8])
  ////
  //for i, argi := range cmds {
  //  if i < 10 {
  //    continue
  //  }
  //  log.Println(i, argi)
  //  Expect(t, strings.HasSuffix(argi, checks[i-2])).ToBe(true)
  //}
  ////p.Env()
  //ex := runner.LocalExecutor{}
  //err = os.RemoveAll("/tmp/testcwl")
  //Expect(t, err).ToBe(nil)
  //pid, ret, err := ex.Run(p)
  //Expect(t, err).ToBe(nil)
  //t.Log(pid)
  //retCode, _ := <-ret
  //Expect(t, retCode).ToBe(0)
  //outputs, err := e.Outputs()
  //Expect(t, err).ToBe(nil)
  //t.Log(outputs)
  //outputRaw := `{"args":["bwa", "mem", "-t", "2", "-I", "1,2,3,4", "-m", "3",
  //   "chr20.fa","example_human_Illumina.pe_1.fastq","example_human_Illumina.pe_2.fastq"]}`
  //outputCheck := cwl.Values{}
  //json.Unmarshal([]byte(outputRaw), &outputCheck)
  //Expect(t, outputs).ToBe(outputCheck)
}
