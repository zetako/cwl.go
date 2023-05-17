package runner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/lijiang2014/cwl.go"
)

// Importer to support $include & $import
// https://www.commonwl.org/v1.0/CommandLineTool.html#Document_preprocessing
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

func (e *Engine) importBeans(bean interface{}) (out interface{}, err error) {
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
			var ret string
			var iret json.RawMessage
			ret, err = e.tryInclude(value)
			if err != nil {
				return nil, err
			}
			if ret != "" {
				t[key] = ret
				continue
			}
			iret, err = e.tryImport(value)
			if err != nil {
				return nil, nil
			}
			if iret != nil {
				var ival interface{}
				if err = json.Unmarshal(iret, &ival); err != nil {
					return nil, err
				}
				t[key] = ival
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
			var ret string
			var iret json.RawMessage
			ret, err = e.tryInclude(value)
			if err != nil {
				return nil, err
			}
			if ret != "" {
				t[i] = ret
				continue
			}
			iret, err = e.tryImport(value)
			if err != nil {
				return nil, err
			}
			if iret != nil {
				var ival interface{}
				if err = json.Unmarshal(iret, &ival); err != nil {
					return nil, err
				}
				t[i] = ival
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

func (e *Engine) tryInclude(bean interface{}) (string, error) {
	if dict, got := bean.(map[string]interface{}); got {
		if value, got := dict["$include"]; got {
			if len(dict) != 1 {
				return "", fmt.Errorf("bad include format")
			}
			valStr, ok := value.(string)
			if ok {
				data, err := e.importer.Load(valStr)
				if err != nil {
					return "", fmt.Errorf("import($include) %s Err : %s", valStr, err)
				}
				return string(data), nil
			}
		}
	}
	return "", nil
}

func (e *Engine) tryImport(bean interface{}) (json.RawMessage, error) {
	if dict, got := bean.(map[string]interface{}); got {
		if value, got := dict["$import"]; got {
			if len(dict) != 1 {
				return nil, fmt.Errorf("bad import format")
			}
			valStr, ok := value.(string)
			if ok {
				data, err := e.importer.Load(valStr)
				if err != nil {
					return nil, fmt.Errorf("import($import) %s Err : %s", valStr, err)
				}
				// 可能为 YAML
				return cwl.Y2J(data)
				// return data, nil
			}
		}
	}
	return nil, nil
}

func (e *Engine) tryImportRun(wfDoc *cwl.Workflow, graphs cwl.Graphs, count int) error {
	// count
	max := e.Flags.MaxWorkflowNested
	if max <= 0 {
		max = DefaultWorkflowNested
	}
	if count > max {
		return fmt.Errorf("import recursive count reached %d", count)
	}
	// each step
	for index, step := range wfDoc.Steps {
		run := step.Run
		if run.Process == nil {
			// get fileName and fragID
			var fileName, fragID string
			if run.ID == "" {
				return fmt.Errorf("no Process or Run.ID to use")
			}
			tmp := strings.IndexByte(run.ID, '#')
			if tmp == -1 {
				fileName = run.ID
			} else {
				fileName = run.ID[:tmp]
				fragID = run.ID[tmp+1:]
			}
			// 3 possible types:
			// a. [file]: read file as process
			// b. #[frag]: use graph to read
			// c. [file]#[frag]: read file as graph, then process as b
			var tmpRoot cwl.Root
			if fileName != "" {
				if !strings.HasSuffix(fileName, ".cwl") {
					return fmt.Errorf("Run.ID not a cwl file")
				}
				cwlFileReader, err := e.importer.Load(fileName)
				if err != nil {
					return err
				}
				cwlFileJSON, err := cwl.Y2J(cwlFileReader)
				if err != nil {
					return err
				}
				cwlFileJSON, err = e.EnsureImportedDoc(cwlFileJSON)
				if err != nil {
					return err
				}
				if err = json.Unmarshal(cwlFileJSON, &tmpRoot); err != nil {
					return err
				}
				graphs = tmpRoot.Graph
			}
			if fragID == "" { // a
				wfDoc.Steps[index].Run.Process = tmpRoot.Process
			} else {
				for _, graph := range graphs {
					if graph.Process.Base().ID == fragID {
						wfDoc.Steps[index].Run.Process = graph.Process
						break
					}
				}
			}
		}
		// if workflow, recursive
		if wfProc, ok := wfDoc.Steps[index].Run.Process.(*cwl.Workflow); ok {
			if err := e.tryImportRun(wfProc, graphs, count+1); err != nil {
				return err
			}
		}
	}
	return nil
}
