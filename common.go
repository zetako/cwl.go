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
