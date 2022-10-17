package cwl_test

import (
	"encoding/json"
	"github.com/lijiang2014/cwl.go"
	"io/ioutil"
	"os"
	"testing"
)

func TestCWL_process_1(t *testing.T) {
	file, err := os.Open("cwl/v1.0/v1.0/bwa-mem-tool.cwl")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = cwl.Y2J(raw)
	if err != nil {
		t.Fatal(err)
	}
	p := cwl.ProcessBase{}
	err = json.Unmarshal(raw, &p)
	if err != nil {
		t.Fatal(err)
	}
	//t.Log(p)
	if p.CWLVersion != "v1.0" {
		t.Fail()
	}
	if len(p.Inputs) != 5 {
		t.Fail()
	}
	in0 := p.Inputs[0].GetInputParameter()
	if in0.ID != "reference" {
		t.Fail()
	}
	if len(p.Outputs) != 2 {
		t.Fail()
	}
	out0 := p.Outputs[0].GetOutputParameter()
	if out0.ID != "sam" {
		t.Fail()
	}
}

func TestCWL_process_2(t *testing.T) {
	p, err := loadProcess("cwl/v1.0/v1.0/cat-tool.cwl")
	if err != nil {
		t.Fatal(err)
	}
	//t.Log(p)
	if p.CWLVersion != "v1.0" {
		t.Fail()
	}
	if len(p.Inputs) != 1 {
		t.Fail()
	}
	in0 := p.Inputs[0].GetInputParameter()
	if in0.ID != "file1" {
		t.Log(in0)
		t.Fail()
	}
	if len(p.Outputs) != 1 {
		t.Fail()
	}
	out0 := p.Outputs[0].GetOutputParameter()
	if out0.ID != "output" {
		t.Log(out0)
		t.Fail()
	}
}

func loadProcess(filename string) (p *cwl.ProcessBase, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	jraw, err := cwl.Y2J(raw)
	if err != nil {
		return nil, err
	}
	p = &cwl.ProcessBase{}
	err = json.Unmarshal(jraw, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

//cwl/v1.0/v1.0/step-valuefrom2-wf.cwl

// Test Requirement parse
func TestCWL_process_3(t *testing.T) {
	p, err := loadProcess("cwl/v1.0/v1.0/step-valuefrom2-wf.cwl")
	if err != nil {
		t.Fatal(err)
	}
	//t.Log(p)
	if p.CWLVersion != "v1.0" {
		t.Fail()
	}
	if len(p.Inputs) != 2 {
		t.Fail()
	}
	in0 := p.Inputs[0].GetInputParameter()
	if in0.ID != "a" && in0.ID != "b" {
		t.Log(in0)
		t.Fail()
	}
	if len(p.Outputs) != 1 {
		t.Fail()
	}
	out0 := p.Outputs[0].GetOutputParameter()
	if out0.ID != "val" {
		t.Log(out0)
		t.Fail()
	}
	if len(p.Requirements) != 3 {
		t.Fail()
	}
	req1 := p.Requirements[1]
	if req1.ClassName() != "InlineJavascriptRequirement" {
		t.Fatal(req1.ClassName())
	}
}
