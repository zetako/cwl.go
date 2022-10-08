package runnertest

import (
  "encoding/json"
  "github.com/otiai10/cwl.go"
  "github.com/otiai10/cwl.go/irunner"
  "github.com/otiai10/cwl.go/runner"
  "io/ioutil"
  "log"
  "os"
  "strings"
  "testing"
  
  . "github.com/otiai10/mint"
)

func TestCWL_basename_fields_test_test(t *testing.T) {
  f := load("v1.0/bwa-mem-tool.cwl")
  req := runner.WorkflowRequest{}
  data1,_ := ioutil.ReadAll(f)
  jd1, _ :=cwl.Y2J(data1)
  req.Workflow = jd1
  f2 := load("v1.0/bwa-mem-job.json")
  
  data2,_ := ioutil.ReadAll(f2)
  jd2, _ :=cwl.Y2J(data2)
  req.Input = jd2
  raw, err := json.Marshal(req)
  Expect(t, err).ToBe(nil)
  err = runner.Engine("test", raw)
  Expect(t, err).ToBe(nil)
  t.Log(err)
}

func TestCWLR2_detail(t *testing.T) {
  e, err := newEngine("v1.0/bwa-mem-tool.cwl", "v1.0/bwa-mem-job.json")
  Expect(t, err).ToBe(nil)
  p, err := e.MainProcess()
  Expect(t, err).ToBe(nil)
  cmds , err := p.Command()
  Expect(t, err).ToBe(nil)
  log.Println(cmds)
  for i, args := range cmds {
    log.Println(i, args)
  }
  Expect(t, len(cmds)).ToBe(13)
  //Expect(t, cmds[3]).ToBe(["bwa","mem"])
  checks := []string{"bwa","mem","-t", "2", "-I", "1,2,3,4", "-m", "3","chr20.fa",
    "example_human_Illumina.pe_1.fastq",
    "example_human_Illumina.pe_2.fastq"}
  ExpectArray(t, cmds[2:10], checks[:8])
  
  for i, argi := range cmds {
    if i < 10 {
      continue
    }
    log.Println(i, argi)
    Expect(t, strings.HasSuffix(argi, checks[i-2])).ToBe(true)
  }
  p.Env()
  ex := irunner.LocalExecutor{}
  err = os.RemoveAll("/tmp/testcwl")
  Expect(t, err).ToBe(nil)
  pid, ret, err := ex.Run(p)
  Expect(t, err).ToBe(nil)
  t.Log(pid)
  retCode ,_ := <- ret
  Expect(t, retCode).ToBe(0)
  outputs , err := e.Outputs()
  Expect(t, err).ToBe(nil)
  t.Log(outputs)
  outputRaw := `{"args":["bwa", "mem", "-t", "2", "-I", "1,2,3,4", "-m", "3",
      "chr20.fa","example_human_Illumina.pe_1.fastq","example_human_Illumina.pe_2.fastq"]}`
  outputCheck := cwl.Values{}
  json.Unmarshal([]byte(outputRaw), &outputCheck)
  Expect(t, outputs).ToBe(outputCheck)
}

func TestCWLR2_run1(t *testing.T) {
  toTests := filterTests(TestDoc{ID: 13})
  Expect(t, len(toTests)).ToBe(1)
  doc := toTests[0]
  doTest(t, doc)
}

