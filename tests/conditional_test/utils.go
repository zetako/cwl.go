package conditionaltest

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/lijiang2014/cwl.go"
	irunner "github.com/lijiang2014/cwl.go/runner"
	. "github.com/otiai10/mint"
)

const version = "1.0"

var (
	tests    []TestDoc
	pathBase string = "/home/zetako/git/cwl-v1.2/tests/conditionals"
)

// Provides file object for testable official .cwl files.
func load(name string) *os.File {
	f, err := os.Open(path.Join(pathBase, name))
	if err != nil {
		panic(err)
	}
	return f
}

func newEngine(tool, param string) (*irunner.Engine, error) {
	var (
		documentID string
	)
	baseTool := path.Base(tool)
	if strings.Contains(baseTool, "#") {
		tmpIdx := strings.IndexByte(baseTool, '#')
		toolName := baseTool[:tmpIdx]
		documentID = baseTool[tmpIdx+1:]
		tool = path.Join(path.Dir(tool), toolName)
	}
	f := load(tool)
	data1, _ := ioutil.ReadAll(f)
	jd1, _ := cwl.Y2J(data1)

	f2 := load(param)
	data2, _ := ioutil.ReadAll(f2)
	jd2, _ := cwl.Y2J(data2)
	return irunner.NewEngine(irunner.EngineConfig{
		DocumentID:   documentID,
		RunID:        "testcwl",
		RootHost:     "/tmp/testcwl/",
		InputsDir:    "inputs",
		WorkDir:      "run",
		Process:      jd1,
		Params:       jd2,
		DocImportDir: pathBase,
	})
}

func ExpectArray(t *testing.T, v []string, wanna []string) {
	Expect(t, len(v)).ToBe(len(wanna))
	for i, _ := range wanna {
		Expect(t, v[i]).ToBe(wanna[i])
	}
}

func Reload(testIndexFile string) error {
	f := load("test-index.yaml")
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	tests = make([]TestDoc, 0)
	bj, err := cwl.Y2J(b)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bj, &tests)
	if err != nil {
		return err
	}
	return nil
}

func SwitchTestSet(path, file string) error {
	if path != pathBase {
		pathBase = path
		return Reload(file)
	} else {
		return nil
	}
}

func init() {
	err := Reload("test-index.yaml")
	if err != nil {
		panic(err)
	}
}

type TestDoc struct {
	ID         string `json:"id"`
	Tags       []string
	Label      string
	Tool       string
	Job        string
	Output     cwl.Values
	Doc        string
	ShouldFail bool `json:"should_fail"`
}

func (recv *TestDoc) UnmarshalJSON(b []byte) error {
	type alias TestDoc
	doc := (*alias)(recv)
	recv.Output = cwl.Values{}
	return json.Unmarshal(b, doc)
}

func filterTests(search TestDoc) []TestDoc {
	var out []TestDoc
	if search.ID == "" && search.Label == "" && len(search.Tags) == 0 {
		return tests
	}
	if search.ID != "" || search.Label != "" {
		for _, testi := range tests {
			if testi.ID == search.ID || testi.Label == search.Label {
				return []TestDoc{testi}
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
	return out
}

func doTest(t *testing.T, doc TestDoc) {
	var rawout []byte
	defer func() {
		if t.Failed() {
			t.Logf("Test Failed: %s %s %s", doc.ID, doc.Tool, doc.Job)
			t.Logf("Labels: %s Tag: %v ", doc.Label, doc.Tags)
			if !doc.ShouldFail {
				t.Logf("actual outraw: %s ", string(rawout))
				rawout, _ = json.Marshal(doc.Output)
				t.Logf("excepted outraw: %s ", string(rawout))
			}

		}
	}()
	e, err := newEngine(doc.Tool, doc.Job)
	Expect(t, err).ToBe(nil)
	if t.Failed() {
		return
	}
	ex := &irunner.LocalExecutor{}
	err = os.RemoveAll("/tmp/testcwl")
	e.SetDefaultExecutor(ex)
	outputs, err := e.Run()
	if !doc.ShouldFail {
		Expect(t, err).ToBe(nil)
		rawout, _ = json.Marshal(outputs)
		//Expect(t, outputs).ToBe(doc.Output)
		if !ExpectOutputs(outputs, doc.Output) {
			Expect(t, outputs).ToBe(doc.Output)
			//t.Fail()
		}
	} else if err == nil {
		Expect(t, err).Not().ToBe(nil)
	}
}

func ExpectOutputs(actual interface{}, expect interface{}) bool {
	switch t := expect.(type) {
	case map[string]interface{}:
		vt := cwl.Values{}
		for k, v := range t {
			vt[k] = v
		}
		return ExpectOutputs(actual, vt)
	case map[string]cwl.Value:
		vt := cwl.Values{}
		for k, v := range t {
			vt[k] = v
		}
		return ExpectOutputs(actual, vt)
	}
	switch t := expect.(type) {
	case cwl.Values:
		amap, ok := actual.(cwl.Values)
		if !ok {
			return false
		}
		for key, val := range t {
			aval := amap[key]
			if !ExpectOutputs(aval, val) {
				return false
			}
		}
		return true
	case []cwl.Value:
		alist, ok := actual.([]cwl.Value)
		if !ok {
			return false
		}
		if len(alist) != len(t) {
			return false
		}
		for i, val := range t {
			aval := alist[i]
			if !ExpectOutputs(aval, val) {
				return false
			}
		}
		return true
	case cwl.Directory:
		aDict, ok := actual.(cwl.Directory)
		if !ok {
			return false
		}
		if t.Location != "Any" && t.Location != "" {
			if !strings.HasSuffix(aDict.Location, t.Location) {
				return false
			}
			// return true
		}
		if len(t.Listing) != len(aDict.Listing) {
			return false
		}
		for i, entryi := range t.Listing {
			if !ExpectOutputs(entryi, t.Listing[i]) {
				return false
			}
		}
		return true
	case cwl.File:
		aFile, ok := actual.(cwl.File)
		if !ok {
			return false
		}
		if t.Location != "Any" && t.Location != "" {
			if !strings.HasSuffix(aFile.Location, t.Location) {
				return false
			}
			// return true
		}
		if len(t.SecondaryFiles) != len(aFile.SecondaryFiles) {
			return false
		}
		for i, val := range t.SecondaryFiles {
			aval := aFile.SecondaryFiles[i]
			if !ExpectOutputs(aval, val) {
				return false
			}
		}
		return (t.Checksum == "" || t.Checksum == aFile.Checksum) &&
			t.Size == aFile.Size && t.Contents == aFile.Contents
	case float64:
		switch at := actual.(type) {
		case float64:
			return t == at
		case int64:
			return t == float64(at)
		case int32:
			return t == float64(at)
		case int:
			return t == float64(at)
		}
		return false
	default:
		return actual == expect
	}
}
