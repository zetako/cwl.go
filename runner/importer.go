package runner

import (
	"encoding/json"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Importer interface {
	Load(string) ([]byte, error)
}


type DefaultImporter struct {
	BaseDir string
}

func (i *DefaultImporter) redirectURI(uri string) string {
	return filepath.Join(i.BaseDir, uri)
}

func (i *DefaultImporter) Load(uri string) ([]byte, error) {
	uri = i.redirectURI(uri)
	fs, err := os.Open(uri)
	if err != nil {
		return nil, err
	}
	defer fs.Close()
	return ioutil.ReadAll(fs)
}

func (e *Engine) loadRoot(uri string) (*cwl.Root, error) {
	data, err := e.importer.Load(uri)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(uri, ".cwl") {
		data, _ = cwl.Y2J(data)
	}
	root := &cwl.Root{}
	err = json.Unmarshal(data, root)
	return root, err
}


func (e *Engine) EnsureImportedDoc(data []byte) ([]byte, error) {
	var bean interface{}
	err := json.Unmarshal(data, &bean)
	if err != nil {
		return nil, err
	}
	bean, err = e.importBeans(bean)
	if err != nil {
		return nil, err
	}
	return json.Marshal(bean)
}

func (e *Engine) importBeans(bean interface{} ) ( out interface{}  ,err error) {
	switch t := bean.(type) {
	case map[string]interface{}:
		ret, err := e.tryInclude(t)
		if err != nil {
			return nil, err
		}
		if ret != "" {
			return ret, nil
		}
		for key, value := range t {
			ret, err := e.tryInclude(value)
			if err != nil {
				return nil, err
			}
			if ret != "" {
				t[key] = ret
				continue
			}
			out, err = e.importBeans(value)
			if err != nil {
				return nil, err
			}
			t[key] = out
		}
		return t, nil
	case []interface{}:
		for i, value := range t {
			ret, err := e.tryInclude(value)
			if err != nil {
				return nil, nil
			}
			if ret != "" {
				t[i] = ret
				continue
			}
			out, err = e.importBeans(value)
			if err != nil {
				return nil, err
			}
			t[i] = out
		}
		return t, nil
	default:
		return bean, nil
	}
}


func (e *Engine) tryInclude(bean interface{} ) ( string , error) {
	if dict, got  := bean.(map[string]interface{}); got {
		if value , got := dict["$include"]; got {
			if len(dict) != 1 {
				return "", fmt.Errorf("bad import format")
			}
			valStr, ok := value.(string)
			if ok {
				data, err := e.importer.Load(valStr)
				if err != nil {
					return "", fmt.Errorf("import %s Err : %s", valStr, err)
				}
				return string(data), nil
			}
		}
	}
	return "", nil
}

func (e *Engine) tryImport(bean interface{} ) ( json.RawMessage , error) {
	if dict, got  := bean.(map[string]interface{}); got {
		if value , got := dict["$include"]; got {
			if len(dict) != 1 {
				return nil, fmt.Errorf("bad import format")
			}
			valStr, ok := value.(string)
			if ok {
				data, err := e.importer.Load(valStr)
				if err != nil {
					return nil, fmt.Errorf("import %s Err : %s", valStr, err)
				}
				return data, nil
			}
		}
	}
	return nil, nil
}
