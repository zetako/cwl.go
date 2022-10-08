package cwl_test

import (
  "github.com/otiai10/cwl.go"
  "io/ioutil"
  "os"
  "testing"
)

func TestCWL_expressionTool_1(t *testing.T) {
  file , err := os.Open("cwl/v1.0/v1.0/null-expression2-tool.cwl")
  if err != nil {
    t.Fatal(err)
  }
  raw, err := ioutil.ReadAll(file)
  if err != nil {
    t.Fatal(err)
  }
  pp , err := cwl.ParseCWLProcess(raw)
  if err != nil {
    t.Fatal(err)
  }
  p , ok := pp.(*cwl.ExpressionTool)
  if !ok {
    t.Fatalf("not workflow : %#v", p)
  }
  
  if p.ClassName() != "ExpressionTool" {
    t.Fatal( "ClassName", p.ClassName())
  }
  pass := true
  pass = p.CWLVersion == "v1.0" &&
    len(p.Inputs) == 1 && p.Inputs[0].GetInputParameter().ID == "i1" &&
    len(p.Outputs) == 1 &&
    p.Outputs[0].GetOutputParameter().ID == "output"
  if !pass {
    t.Fatalf("%#v",p)
  }
  i1,ok1 := p.Inputs[0].(*cwl.WorkflowInputParameter)
  o1, ok2 := p.Outputs[0].(*cwl.ExpressionToolOutputParameter)
  
  pass = ok1 && ok2 &&
    i1.Type.TypeName() == "Any" &&
    o1.Type.TypeName() == "int"
  if !pass {
    t.Fatalf("%#v %#v",p.Inputs, p.Outputs)
  }
  
  // ✅ Test Requirements
  reqs := p.Requirements
  reqs0, ok0 := reqs[0].(*cwl.InlineJavascriptRequirement)
  pass = len(reqs) == 1 && ok0  &&
    reqs0.ClassName() == "InlineJavascriptRequirement"
  if !pass {
    t.Fatalf("Requirements check failed %#v",p.Requirements)
  }

}

func TestCWL_workflow_1(t *testing.T) {
  
  file , err := os.Open("cwl/v1.0/v1.0/basename-fields-test.cwl")
  if err != nil {
    t.Fatal(err)
  }
  raw, err := ioutil.ReadAll(file)
  if err != nil {
    t.Fatal(err)
  }
  pp , err := cwl.ParseCWLProcess(raw)
  if err != nil {
   t.Fatal(err)
  }
  p , ok := pp.(*cwl.Workflow)
  if !ok {
    t.Fatalf("not workflow : %#v", p)
  }
  if p.ClassName() != "Workflow" {
    t.Fatal( "ClassName", p.ClassName())
  }
  pass := true
  pass = p.CWLVersion == "v1.0" &&
    len(p.Inputs) == 1 && p.Inputs[0].GetInputParameter().ID == "tool" &&
    len(p.Outputs) == 2
  if !pass {
    t.Fatalf("%#v",p)
  }
  
  // ✅ Test Requirements
  reqs := p.Requirements
  reqs0, ok0 := reqs[0].(*cwl.StepInputExpressionRequirement)
  pass = len(reqs) == 1 && ok0  &&
    reqs0.ClassName() == "StepInputExpressionRequirement"
  if !pass {
    t.Fatalf("Requirements check failed %#v",p.Requirements)
  }
  // ✅ Test Inputs
  in1 := p.Inputs[0].(*cwl.WorkflowInputParameter)
  pass = in1.Type.TypeName() == "File"
  if !pass {
    t.Fatalf("err input %#v",in1)
  }
  // ✅ Test Outputs
  out1 := p.Outputs[0].(*cwl.WorkflowOutputParameter)
  out2 := p.Outputs[1].(*cwl.WorkflowOutputParameter)
  
  if out2.ID == "rootFile" {
    out1, out2 = out2, out1
  }
  pass = out1.ID == "rootFile" && out2.ID == "extFile" &&
    out1.OutputSource[0] == "root/out" && out1.Type.TypeName() == "File" &&
    out2.OutputSource[0] == "ext/out" && out1.Type.TypeName() == "File"
  if !pass {
    t.Fatalf("err outputs %#v %#v",out1, out2)
  }
  // ✅  Test Steps
  pass = len(p.Steps) == 2
  if !pass {
    t.Fatalf("err steps %#v",p.Steps)
  }
  step1 := p.Steps[0]
  step2 := p.Steps[1]
  if step2.ID == "root" {
    step1, step2 = step2, step1
  }
  pass = step1.ID == "root" && step1.Run.ID == "echo-file-tool.cwl" &&
   len(step1.In) == 2 && len(step1.Out) == 1 && step1.Out[0].ID == "out" &&
  step2.ID == "ext" && step2.Run.ID == "echo-file-tool.cwl" &&
   len(step2.In) == 2 && len(step2.Out) == 1 && step2.Out[0].ID == "out"
  if !pass {
    t.Fatalf("err steps %#v \n %#v" ,step1 , step2)
  }
  s1in1 := step1.In[0]
  s1in2 := step1.In[1]
  if s1in2.ID == "tool" {
    s1in1, s1in2 = s1in2, s1in1
  }
  pass = s1in1.ID == "tool" && s1in1.Source[0] == "tool" &&
    s1in2.ID == "in" && s1in2.ValueFrom == "$(inputs.tool.nameroot)"
  if !pass {
    t.Fatalf("err step in %#v %#v" ,s1in1, s1in2)
  }
  s2in1 := step2.In[0]
  s2in2 := step2.In[1]
  if s2in2.ID == "tool" {
    s2in1, s2in2 = s2in2, s2in1
  }
  pass = s2in1.ID == "tool" && s2in1.Source[0] == "tool" &&
    s2in2.ID == "in" && s2in2.ValueFrom == "$(inputs.tool.nameext)"
  if !pass {
    t.Fatalf("err step in %#v %#v" ,s2in1, s2in2)
  }
}