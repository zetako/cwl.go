package cwl_test

import (
  "encoding/json"
  "github.com/lijiang2014/cwl.go"
  "io/ioutil"
  "os"
  "testing"
)

func TestCWL_value_1(t *testing.T) {
  file, err := os.Open("cwl/v1.0/v1.0/bwa-mem-job.json")
  if err != nil {
    t.Fatal(err)
  }
  data,  _ := ioutil.ReadAll(file)
  values := cwl.Values{}
  err = json.Unmarshal(data, &values)
  if err != nil {
    t.Fatal(err)
  }
  t.Log(values)
}

// {"class":"File","location":"args.py"}

func TestCWL_value_2(t *testing.T) {
  //file, err := os.Open("cwl/v1.0/v1.0/bwa-mem-job.json")
  //if err != nil {
  //  t.Fatal(err)
  //}
  data := `{"in":{"class":"File","location":"args.py"}}`
  values := cwl.Values{}
  err := json.Unmarshal([]byte(data), &values)
  if err != nil {
    t.Fatal(err)
  }
  t.Log(values, values["in"].(cwl.File))
}

//   fieldValue.Set(reflect.MakeSlice(reflect.SliceOf(fieldType), len(values), len(values)))
func TestCWL_value_3(t *testing.T) {
  values := cwl.Values{"list": nil}
  //listV.SetCap(2)
  //listV.SetLen(2)
  
  data := `{"list":[{"class":"File","location":"args.py"},{"class":"File","location":"args2.py"}]}`
  //values := cwl.Values{}
  err := json.Unmarshal([]byte(data), &values)
  if err != nil {
    t.Fatal(err)
  }
  t.Log(values, values["list"])
}
