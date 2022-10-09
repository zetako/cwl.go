package cwl

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"

	"github.com/robertkrimen/otto"
)

// NewCWL ...
func NewCWL() *Root {
	root := new(Root)
	return root
}

type SaladRootDoc struct {
	Graphs       Graphs `json:"$graphs"`
	
}

// Root ...
type Root struct {
	SaladRootDoc `json:",inline"`
	Process Process
	// Use $graphs 组合
	Class        string       `json:"class"`
	// Path
	Path string `json:"-"`
	// InputsVM
	InputsVM *otto.Otto
}

// UnmarshalJSON ...
func (root *Root) UnmarshalJSON(b []byte) error {
	p, err :=ParseCWLProcess(b)
	if err != nil {
		return err
	}
	root.Process = p
	if c, ok := p.(Classable); ok {
		root.Class = c.ClassName()
	}
	return nil
}

// Decode decodes specified file to this root
func (root *Root) Decode(r io.Reader) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Parse error: %v", e)
		}
	}()
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	buf2, err := Y2J(buf)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(buf2, root); err != nil {
		return err
	}
	return nil
}

//// AsStep constructs Root as a step of "steps" from interface.
//func (root *Root) AsStep(i interface{}) *Root {
//	dest := new(Root)
//	switch x := i.(type) {
//	case string:
//		dest.ID = x
//	case map[string]interface{}:
//		err := dest.UnmarshalMap(x)
//		if err != nil {
//			panic(fmt.Sprintf("Failed to parse step as CWL.Root: %v", err))
//		}
//	}
//	return dest
//}

// Y2J converts yaml to json.
func Y2J(in []byte) ([]byte, error) {
	result := []byte{}
	var root interface{}
	if err := yaml.Unmarshal(in, &root); err != nil {
		return result, err
	}
	return json.Marshal(convert(root))
}

// J2Y converts json to yaml.
func J2Y(r io.Reader) ([]byte, error) {
	result := []byte{}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return result, err
	}
	var root interface{}
	if err := json.Unmarshal(b, &root); err != nil {
		return result, err
	}
	return yaml.Marshal(convert(root))
}

// convert ...
func convert(parent interface{}) interface{} {
	switch entity := parent.(type) {
	case map[interface{}]interface{}:
		node := map[string]interface{}{}
		for key, val := range entity {
			node[key.(string)] = convert(val)
		}
		return node
	case []interface{}:
		for idx, val := range entity {
			entity[idx] = convert(val)
		}
		return entity
	}
	return parent
}
