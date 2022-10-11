package cwl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	
	yaml "gopkg.in/yaml.v2"
)

// Value ...
type Value interface{}

// Values represents specific parameters to run workflow which is described by CWL.
type Values map[string]Value

// NewValues ...
func NewValues() *Values {
	return &Values{}
}

// Decode ...
func (p *Values) Decode(f *os.File) error {
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
	NoClass   ParameterClass = "unknown"
)

func (recv *Values) UnmarshalJSON(b []byte) error {
	if recv == nil {
		recv = NewValues()
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
		v, err := ConvertToValue(value)
		if err != nil {
			return err
		}
		(*recv)[key] = v
	}
	return nil
}

func ConvertToValue(bean interface{}) (out Value, err error) {
	switch t := bean.(type) {
	case []interface{}:
		arr := make([]Value, len(t))
		for i, item := range t {
			v, err := ConvertToValue(item)
			if err != nil {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	case map[string]interface{}:
		tClass, got := t["class"]
		if !got {
			break
		}
		switch tClass {
		case "File":
			var entry File
			raw, err := json.Marshal(bean)
			if err != nil {
				return nil, err
			}
			if err = json.Unmarshal(raw, &entry); err != nil {
				return nil, err
			}
			return entry, nil
		case "Directory":
			var entry Directory
			entry.Listing = make([]FileDir,0)
			raw, err := json.Marshal(bean)
			if err != nil {
				return nil, err
			}
			//if err = json.Unmarshal(raw, &entry); err != nil {
			//	return err
			//}
			err = parseObject( reflect.TypeOf(entry), reflect.ValueOf(&entry), raw, map[string]*RecordFieldGraph{
				"File" : &RecordFieldGraph{Example: File{}},
				"Directory" : &RecordFieldGraph{Example: Directory{}},
			} )
			
			return entry, err
		}
	default:
		return bean, nil
	}
	return bean, nil
}
