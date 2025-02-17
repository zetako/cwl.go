package cwl

import (
	"encoding/json"
	"fmt"
)

type Classable interface {
	ClassName() string
}

type ClassBase struct {
	Class string `json:"class"`
}

func (c ClassBase) ClassName() string {
	return c.Class
}

type Expression string

type ArrayString []string

type IntExpression struct {
	Expression
	Int *int64
}

func (e IntExpression) Value() (int64, bool) {
	if e.Int != nil {
		return *e.Int, true
	}
	return 0, false
}

func (e IntExpression) MustInt() int {
	if e.Int != nil {
		return int(*e.Int)
	}
	return 0
}

type LongFloatExpression struct {
	Expression
	Long  *int64
	Float *float64
}

type LongExpression struct {
	Expression
	Long *int64
}

type BoolExpression struct {
	Expression
	Bool *bool
}

func (s *ArrayString) UnmarshalJSON(data []byte) error {
	ss := make([]string, 0)
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		if err := json.Unmarshal(data, &ss); err != nil {
			return err
		}
		*s = append(*s, ss...)
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = append(*s, str)
	return nil
}

// StringArrayable converts "xxx" to ["xxx"] if it's not slice.
func StringArrayable(i interface{}) []string {
	dest := []string{}
	switch x := i.(type) {
	case []interface{}:
		for _, s := range x {
			dest = append(dest, s.(string))
		}
	case string:
		dest = append(dest, x)
	}
	return dest
}

type ArrayExpression []Expression

func (s *ArrayExpression) UnmarshalJSON(data []byte) error {
	ss := make([]Expression, 0)
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		if err := json.Unmarshal(data, &ss); err != nil {
			return err
		}
		*s = append(*s, ss...)
		return nil
	}
	var str Expression
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = append(*s, str)
	return nil
}

func (e *IntExpression) UnmarshalJSON(data []byte) error {
	var bean interface{}
	err := json.Unmarshal(data, &bean)
	if err != nil {
		return err
	}
	switch v := bean.(type) {
	case string:
		e.Expression = Expression(v)
		return nil
	case float64:
		// 直接使用可能有精度损失
		var num int64
		err := json.Unmarshal(data, &num)
		if err != nil {
			return err
		}
		e.Int = &num
		return nil
	}
	return fmt.Errorf("only int/Expression is available")
}

func (e *LongFloatExpression) UnmarshalJSON(data []byte) error {
	var bean interface{}
	err := json.Unmarshal(data, &bean)
	if err != nil {
		return err
	}
	switch v := bean.(type) {
	case string:
		e.Expression = Expression(v)
		return nil
	case float64:
		// 通过正则来判断可能更合适？
		// 直接使用可能有精度损失
		var num int64
		err := json.Unmarshal(data, &num)
		if err != nil {
			return err
		}
		if fmt.Sprint(num) == string(data) {
			e.Long = &num
			return nil
		}
		e.Float = &v
		return nil
	}
	return fmt.Errorf("only long/float/Expression is available")
}

func (e *LongFloatExpression) IsNull() bool {
	return e == nil ||
		(e.Long == nil &&
			e.Float == nil &&
			e.Expression == "")
}

type JavaScriptInterpreter interface {
	Eval(e Expression, data interface{}) (interface{}, error)
}

func (e Expression) Eval(i JavaScriptInterpreter, data interface{}) (interface{}, error) {
	return i.Eval(e, data)
}

func (e *LongFloatExpression) MustFloat() float64 {
	if e.Float != nil {
		return *e.Float
	}
	if e.Long != nil {
		return float64(*e.Long)
	}
	return 0
}

func (e *LongFloatExpression) MustInt64() int64 {
	if e.Long != nil {
		return *e.Long
	}
	if e.Float != nil {
		return int64(*e.Float)
	}
	return 0
}

func (e *LongFloatExpression) Resolve(i JavaScriptInterpreter, data interface{}) error {
	if e.Long != nil || e.Float != nil {
		return nil
	}
	if e.Expression == "" {
		return fmt.Errorf("no Expression")
	}
	out, err := e.Expression.Eval(i, data)
	if err != nil {
		return err
	}
	switch v := out.(type) {
	case int64:
		e.Long = &v
		break
	case float64:
		e.Float = &v
		break
	default:
		return fmt.Errorf("need to be a number")
	}
	return nil
}

func (e *FileDirExpDirent) UnmarshalJSON(data []byte) error {
	var bean interface{}
	err := json.Unmarshal(data, &bean)
	if err != nil {
		return err
	}
	switch v := bean.(type) {
	case string:
		e.Expression = Expression(v)
		return nil
	case map[string]interface{}:
		if v["class"] == "File" {
			e.File = &File{}
			return json.Unmarshal(data, e.File)
		} else if v["class"] == "Directory" {
			e.Directory = &Directory{}
			return json.Unmarshal(data, e.Directory)
		}
		e.Dirent = &Dirent{}
		return json.Unmarshal(data, e.Dirent)
	}
	return fmt.Errorf("only Expression/File/Directory/Dirent is available")
}

func (list *FileDirExpDirentList) UnmarshalJSON(data []byte) error {
	ss := make([]FileDirExpDirent, 0)
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		if err := json.Unmarshal(data, &ss); err != nil {
			return err
		}
		*list = append(*list, ss...)
		return nil
	}
	var filedir FileDirExpDirent
	if err := json.Unmarshal(data, &filedir); err != nil {
		return err
	}
	*list = append(*list, filedir)
	return nil
}

func (e *SecondaryFileSchema) UnmarshalJSON(data []byte) error {
	var bean interface{}
	err := json.Unmarshal(data, &bean)
	if err != nil {
		return err
	}
	e.Required = true
	switch v := bean.(type) {
	case string:
		e.Pattern = v
		return nil
	case map[string]interface{}:
		if pattern, ok := v["pattern"].(string); ok {
			e.Pattern = pattern
		} else {
			return fmt.Errorf("SecondaryFileSchema need pattern")
		}
		if req := v["required"]; req != nil {
			if reqBool, ok := req.(bool); ok {
				e.Required = reqBool
			} else {
				return fmt.Errorf("SecondaryFileSchema.required need be bool")
			}
		}
	default:
		return fmt.Errorf("SecondaryFileSchema type error")
	}
	return nil
}
