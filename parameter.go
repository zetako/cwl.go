package cwl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// Parameter ...
type Parameter interface{}

// Parameters represents specific parameters to run workflow which is described by CWL.
type Parameters map[string]Parameter

// NewParameters ...
func NewParameters() *Parameters {
	return &Parameters{}
}

// Decode ...
func (p *Parameters) Decode(f *os.File) error {
	switch filepath.Ext(f.Name()) {
	case "json":
		return json.NewDecoder(f).Decode(p)
	default:
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(b, p)
	}
}

type ParameterClass string

const (
	FileClass ParameterClass = "File"
	NoClass ParameterClass = "unknown"
)

//func (recv Parameter) Class() ParameterClass {
//	obj, ok := recv.(map[string]interface{})
//	if ok {
//		tClass, got := obj["class"]
//		if !got {
//			return
//		}
//		switch tClass {
//		case "File":
//			var entry Entry
//			json.Unmarshal(b,&entry)
//			recv = &entry
//		}
//	}
//	switch t := recv.(type) {
//	case map[string]interface{}:
//		tClass, got := t["class"]
//		if !got {
//			break
//		}
//		switch tClass {
//		case "File":
//			var entry Entry
//			json.Unmarshal(b,&entry)
//			recv = &entry
//
//		}
//	}
//	return root.UnmarshalMap(docs)
//}

func (recv *Parameters) UnmarshalJSON(b []byte) error {
	if recv == nil {
		recv = NewParameters()
	}
	var any interface{}
	if err := json.Unmarshal(b, &any); err != nil {
		return err
	}
	params, ok := any.(map[string]interface{})
	if !ok {
		return fmt.Errorf("not a key-value type")
	}
	for key, value := range params {
		//switch t := value.(type) {
		//case []interface{}:
		//	arr := make([]Parameter,len(t))
		//	for i, item := range t {
		//		arr[i] = convertParameter(item)
		//	}
		//	(*recv)[key] = Parameter(arr)
		//case map[string]interface{}:
		//	tClass, got := t["class"]
		//	if !got {
		//		break
		//	}
		//	switch tClass {
		//	case "File":
		//		var entry Entry
		//		raw, err := json.Marshal(value)
		//		if err != nil {
		//			return err
		//		}
		//		if err = json.Unmarshal(raw,&entry); err != nil {
		//			return err
		//		}
		//		(*recv)[key] = Parameter(entry)
		//		continue
		//	}
		//}
		(*recv)[key] = convertParameter(value)
	}
	return nil
}

func convertParameter(bean interface{}) (out Parameter) {
	switch t := bean.(type) {
	case []interface{}:
		arr := make([]Parameter,len(t))
		for i, item := range t {
			arr[i] = convertParameter(item)
		}
		return arr
	case map[string]interface{}:
		tClass, got := t["class"]
		if !got {
			break
		}
		switch tClass {
		case "File", "Directory":
			var entry Entry
			raw, err := json.Marshal(bean)
			if err != nil {
				return err
			}
			if err = json.Unmarshal(raw,&entry); err != nil {
				return err
			}
			return entry
		}
	default:
		return bean
	}
	return bean
}