package runnertest

import (
	"encoding/json"
	"fmt"
	"github.com/otiai10/cwl.go"
	"github.com/otiai10/cwl.go/irunner"
	. "github.com/otiai10/mint"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const version = "1.0"

// Provides file object for testable official .cwl files.
func load(name string) *os.File {
	fpath := fmt.Sprintf("../../cwl/v%[1]s/%s", version, name)
	f, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	return f
}



func newEngine(tool , param string) ( *irunner.Engine ,error) {
	f := load(tool)
	data1,_ := ioutil.ReadAll(f)
	jd1, _ :=cwl.Y2J(data1)
	
	f2 := load(param)
	data2,_ := ioutil.ReadAll(f2)
	jd2, _ :=cwl.Y2J(data2)
	
	wd ,_ :=os.Getwd()
	wd = filepath.Join(wd, "../../cwl/v1.0/v1.0")
	return irunner.NewEngine(irunner.EngineConfig{
		RunID: "testcwl",
		RootHost: "/tmp/testcwl",
		Process: jd1,
		Params: jd2,
		Workdir: wd,
	})
}

func ExpectArray(t *testing.T, v []string, wanna []string)  {
	Expect(t, len(v)).ToBe(len(wanna))
	for i, _ := range wanna {
		Expect(t, v[i]).ToBe(wanna[i])
	}
}

var (
	tests []TestDoc
)

func init()  {
	f := load("conformance_test_v1.0.yaml")
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	tests = make([]TestDoc,0)
	bj , err := cwl.Y2J(b)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bj,&tests)
	if err != nil {
		panic(err)
	}
}


type TestDoc struct {
	ID int `json:"id"`
	Tags []string
	Label string
	Tool string
	Job string
	Output cwl.Parameters
	Doc string
}

func (recv *TestDoc) UnmarshalJSON(b []byte) error {
	type alias TestDoc
	doc := (*alias)(recv)
	recv.Output = cwl.Parameters{}
	return json.Unmarshal(b, doc)
}

func filterTests(search TestDoc) []TestDoc {
	var out []TestDoc
	if search.ID == 0 && search.Label == "" && len(search.Tags) == 0 {
		return tests
	}
	if search.ID != 0 || search.Label != "" {
		for _, testi := range tests {
			if testi.ID == search.ID || testi.Label == search.Label  {
				return []TestDoc{ testi }
			}
		}
	}
	for i, testi := range tests {
		for _, tagi := range testi.Tags {
			if tagi == search.Tags[0] {
				out = append(out, tests[i])
			}
		}
	}
	return  out
}

func doTest(t *testing.T, doc TestDoc) {
	e, err := newEngine(doc.Tool, doc.Job)
	Expect(t, err).ToBe(nil)
	p, err := e.MainProcess()
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
	Expect(t, outputs).ToBe(doc.Output)
}
